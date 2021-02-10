import * as R from 'ramda';
import Boom from '@hapi/boom';
import { Group } from '../../model/group';
import * as PolicyResource from '../policy/resource';
import { PolicyOperation } from '../policy/resource';
import CasbinSingleton from '../../lib/casbin';

type JSObj = Record<string, unknown>;

export const appendGroupIdWithPolicies = (
  policies: PolicyOperation[],
  groupId: string
) => {
  return policies.map((policy) => {
    const NAMEPATHS_TO_CHECK = [
      ['subject', 'group'],
      ['resource', 'group']
    ];

    return NAMEPATHS_TO_CHECK.reduce(
      (policyWithGroupId, namepath) => {
        if (R.hasPath(namepath, policy)) {
          return R.assocPath(namepath, groupId, policyWithGroupId);
        }
        return policyWithGroupId;
      },
      { ...policy }
    );
  });
};

export const checkSubjectHasAccessToCreateAttributesMapping = async (
  subject: JSObj,
  attributes: JSObj[]
) => {
  const IAM_ACTION = 'iam.create';
  if (!R.isEmpty(attributes)) {
    const results = await Promise.all(
      attributes.map(async (attribute: JSObj) => {
        return CasbinSingleton?.enforcer?.enforceJson(subject, attribute, {
          action: IAM_ACTION
        });
      })
    );
    if (results.includes(false)) {
      throw Boom.forbidden("Sorry you don't have access");
    }
  }
  return true;
};

export const upsertGroupAndAttributesMapping = async (
  groupId: string,
  attributes: JSObj[]
) => {
  if (R.isEmpty(attributes)) {
    return;
  }

  await CasbinSingleton?.enforcer?.removeAllResourceGroupingJsonPolicy({
    group: groupId
  });
  await Promise.all(
    attributes.map(async (attribute: JSObj) => {
      await CasbinSingleton?.enforcer?.addResourceGroupingJsonPolicy(
        { group: groupId },
        attribute
      );
    })
  );
};

export const bulkUpsertPoliciesForGroup = async (
  groupId: string,
  policies = [],
  loggedInUserId: string
) => {
  const policiesWithGroupId = appendGroupIdWithPolicies(policies, groupId);
  return PolicyResource.bulkOperation(policiesWithGroupId, {
    user: loggedInUserId
  });
};

export const checkSubjectHasAccessToEditGroup = async (
  group: JSObj,
  attributes: JSObj[],
  loggedInUserId: string
) => {
  const groupId = <string>group.id;
  const prevAttributes = <JSObj[]>group.attributes;

  // ? We need to check this only if the attributes change
  if (!R.equals(attributes, prevAttributes)) {
    await checkSubjectHasAccessToCreateAttributesMapping(
      {
        user: loggedInUserId
      },
      [...attributes, ...prevAttributes]
    );
    await upsertGroupAndAttributesMapping(groupId, attributes);
  }

  // ? the user needs access to the group if they need to edit it
  const hasGroupAccess = await CasbinSingleton?.enforcer?.enforceJson(
    { user: loggedInUserId },
    { group: groupId },
    { action: 'iam.manage' }
  );

  if (!hasGroupAccess) {
    throw Boom.forbidden("Sorry you don't have access");
  }
};

export const list = async (filters?: JSObj) => {
  return PolicyResource.getSubjecListWithPolicies('group', filters);
};

export const get = async (groupId: string, filters?: JSObj) => {
  const group = await Group.findOne(groupId);
  if (!group) throw Boom.notFound('group not found');
  const subject = { group: group?.id };
  const policies = await PolicyResource.getPoliciesBySubject(subject, filters);
  const attributes = await PolicyResource.getAttributesForGroup(groupId);
  return { ...group, policies, attributes };
};

export const create = async (payload: any, loggedInUserId: string) => {
  const { policies = [], attributes = [], ...groupPayload } = payload;
  await checkSubjectHasAccessToCreateAttributesMapping(
    {
      user: loggedInUserId
    },
    attributes
  );

  const groupResult = await Group.save(groupPayload);
  const groupId = groupResult.id;

  await upsertGroupAndAttributesMapping(groupId, attributes);

  const policyOperationResult = await bulkUpsertPoliciesForGroup(
    groupId,
    policies,
    loggedInUserId
  );
  const updatedGroup = await get(groupId);
  return { ...updatedGroup, policyOperationResult };
};

export const update = async (
  groupId: string,
  payload: any,
  loggedInUserId: string
) => {
  const { policies = [], attributes = [], ...groupPayload } = payload;
  const groupWithExtraKeys = await get(groupId);
  const group = R.omit(['policies', 'attributes'], groupWithExtraKeys);

  // ? We need to check this only if the attributes change
  await checkSubjectHasAccessToEditGroup(
    groupWithExtraKeys,
    attributes,
    loggedInUserId
  );

  await Group.save({ ...group, ...groupPayload });

  const policyOperationResult = await bulkUpsertPoliciesForGroup(
    groupId,
    policies,
    loggedInUserId
  );
  const updatedGroup = await get(groupId);
  return { ...updatedGroup, policyOperationResult };
};
