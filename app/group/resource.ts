import * as R from 'ramda';
import Boom from '@hapi/boom';
import { createQueryBuilder } from 'typeorm';
import { Group } from '../../model/group';
import * as PolicyResource from '../policy/resource';
import { PolicyOperation } from '../policy/resource';
import { toLikeQuery } from '../policy/util';
import CasbinSingleton from '../../lib/casbin';
import { parseGroupListResult } from './util';
import getUniqName, { validateUniqName } from '../../lib/getUniqName';

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

// ? /api/groups?entity=gojek&user_role=a4343590
// ? 1) Check whether entity attribute and group mapping exists
// ? 2) Count all members of a group
// ? 3) Find all users with specified user_role as well
// ? 4) Check whether current logged in user is mapped with the group
export const list = async (filters: JSObj = {}, loggedInUserId: string) => {
  const { user_role = '', ...resource } = filters;

  const GET_GROUP_DOC = `JSON_AGG(DISTINCT groups.*) AS group_arr`;
  const MEMBER_COUNT = `SUM(CASE WHEN casbin_rule.ptype = 'g' THEN 1 ELSE 0 END) AS member_count`;
  const ATTRIBUTES_AGGREGATE = `JSON_AGG(casbin_rule.*) FILTER (WHERE casbin_rule.ptype = 'g2') AS raw_attributes`;
  const IS_LOGGEDIN_USER_MEMBER = `BOOL_OR(CASE WHEN casbin_rule.ptype = 'g' AND casbin_rule.v0 = '{"user":"${loggedInUserId}"}' THEN true ELSE false END) AS is_member`;

  const JOIN_WITH_RESOURCE_ATTRIBUTES_MAPPING = `casbin_rule.ptype = 'g2' AND casbin_rule.v0 like '%"' || groups.id || '"%'`;
  const JOIN_WITH_USER_MAPPING = `casbin_rule.ptype = 'g' AND casbin_rule.v1 like '%"' || groups.id || '"%'`;
  const JOIN_WITH_POLICIES = `casbin_rule.ptype = 'p' AND casbin_rule.v1 like '%"' || groups.id || '"%'`;

  const roleQuery = toLikeQuery({ role: user_role });
  const userQuery = `%user":%`;
  const AGGREGATE_MEMBER_POLICIES = R.isEmpty(user_role)
    ? ''
    : `, JSON_AGG(casbin_rule.*) FILTER (WHERE casbin_rule.ptype = 'p' AND casbin_rule.v0 like '${userQuery}' AND casbin_rule.v2 like '${roleQuery}') AS raw_user_policies`;

  const cursor = createQueryBuilder()
    .select(
      `groups.id, ${GET_GROUP_DOC}, ${ATTRIBUTES_AGGREGATE}, ${MEMBER_COUNT}, ${IS_LOGGEDIN_USER_MEMBER} ${AGGREGATE_MEMBER_POLICIES}`
    )
    .from(Group, 'groups')
    .leftJoin(
      'casbin_rule',
      'casbin_rule',
      `(${JOIN_WITH_RESOURCE_ATTRIBUTES_MAPPING}) OR (${JOIN_WITH_USER_MAPPING}) OR (${JOIN_WITH_POLICIES})`
    )
    .groupBy('groups.id');

  if (!R.isEmpty(resource)) {
    const FILTER_BY_RESOURCE_ATTRIBUTES = `SUM(CASE WHEN casbin_rule.ptype = 'g2' AND casbin_rule.v1 like :attribute THEN 1 ELSE 0 END) > 0`;
    cursor.having(FILTER_BY_RESOURCE_ATTRIBUTES, {
      attribute: toLikeQuery(resource)
    });
  }

  const rawResult = await cursor.getRawMany();
  return parseGroupListResult(rawResult);
};

export const get = async (groupId: string, filters?: JSObj) => {
  const group = await Group.findOne(groupId);
  if (!group) throw Boom.notFound('group not found');
  const subject = { group: group?.id };
  const policies = await PolicyResource.getPoliciesBySubject(subject, filters);
  const attributes = await PolicyResource.getAttributesForGroup(groupId);
  return { ...group, policies, attributes };
};

const getValidGroupname = async (payload: any) => {
  let groupname = payload?.groupname;
  if (payload?.displayname && !groupname) {
    groupname = await getUniqName(payload?.displayname, 'groupname', Group);
  }
  validateUniqName(groupname);
  return groupname;
};

export const create = async (payload: any, loggedInUserId: string) => {
  const { policies = [], attributes = [], ...groupPayload } = payload;
  await checkSubjectHasAccessToCreateAttributesMapping(
    {
      user: loggedInUserId
    },
    attributes
  );

  const groupname = await getValidGroupname(groupPayload);
  const groupResult = await Group.save({ ...groupPayload, groupname });
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
