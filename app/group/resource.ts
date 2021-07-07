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
import { User } from '../../model/user';

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
  attributes: JSObj[],
  loggedInUser: User | undefined
) => {
  if (R.isEmpty(attributes)) {
    return;
  }
  const options = { created_by: loggedInUser };
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  await CasbinSingleton?.enforcer?.removeAllResourceGroupingJsonPolicy(
    {
      group: groupId
    },
    options
  );
  await Promise.all(
    attributes.map(async (attribute: JSObj) => {
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      await CasbinSingleton?.enforcer?.addResourceGroupingJsonPolicy(
        { group: groupId },
        attribute,
        options
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
  const user = await User.findOne({
    where: {
      id: loggedInUserId
    }
  });
  // ? We need to check this only if the attributes change
  if (!R.equals(attributes, prevAttributes)) {
    await checkSubjectHasAccessToCreateAttributesMapping(
      {
        user: loggedInUserId
      },
      [...attributes, ...prevAttributes]
    );
    await upsertGroupAndAttributesMapping(groupId, attributes, user);
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
export const list = async (filters: JSObj = {}, loggedInUserId = '') => {
  const { user_role = '', group, ...attributes } = filters;

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

  // ? this is to filter single group query if passed in filters
  if (!R.isNil(group)) {
    cursor.where(`groups.id = :groupId`, { groupId: group });
  }

  if (!R.isEmpty(attributes)) {
    const FILTER_BY_RESOURCE_ATTRIBUTES = `SUM(CASE WHEN casbin_rule.ptype = 'g2' AND casbin_rule.v1 like :attribute THEN 1 ELSE 0 END) > 0`;
    cursor.having(FILTER_BY_RESOURCE_ATTRIBUTES, {
      attribute: toLikeQuery(attributes)
    });
  }

  const rawResult = await cursor.getRawMany();
  return parseGroupListResult(rawResult);
};

export const get = async (
  groupId: string,
  loggedInUserId = '',
  filters: JSObj = {}
) => {
  const filtersWithoutFields = R.omit(['fields'], filters);
  const groupList = await list(
    { group: groupId, ...filtersWithoutFields },
    loggedInUserId
  );
  const group = R.head(groupList);
  if (!group) throw Boom.notFound('group not found');
  const subject = { group: R.path(['id'], group) };
  const policies = await PolicyResource.getPoliciesBySubject(subject, filters);
  return { ...group, policies };
};

const getValidGroupname = async (payload: any) => {
  let groupname = payload?.groupname;
  if (payload?.displayname && !groupname) {
    groupname = await getUniqName(payload?.displayname, 'groupname', Group);
  }
  validateUniqName(groupname);
  return groupname.toLowerCase();
};

export const create = async (payload: any, loggedInUserId: string) => {
  const { policies = [], attributes = [], ...groupPayload } = payload;
  await checkSubjectHasAccessToCreateAttributesMapping(
    {
      user: loggedInUserId
    },
    attributes
  );

  const user = await User.findOne({
    where: {
      id: loggedInUserId
    }
  });
  const groupname = await getValidGroupname(groupPayload);
  const groupResult = await Group.save(
    { ...groupPayload, groupname },
    { data: { user } }
  );
  const groupId = groupResult.id;

  await upsertGroupAndAttributesMapping(groupId, attributes, user);

  const policyOperationResult = await bulkUpsertPoliciesForGroup(
    groupId,
    policies,
    loggedInUserId
  );
  const updatedGroup = await get(groupId, loggedInUserId);
  return { ...updatedGroup, policyOperationResult };
};

export const update = async (
  groupId: string,
  payload: any,
  loggedInUserId: string
) => {
  const { policies = [], attributes = [], ...groupPayload } = payload;
  const groupWithExtraKeys = await get(groupId, loggedInUserId);
  const group = R.omit(['policies', 'attributes'], groupWithExtraKeys);
  const user = await User.findOne({
    where: {
      id: loggedInUserId
    }
  });

  // ? We need to check this only if the attributes change
  await checkSubjectHasAccessToEditGroup(
    groupWithExtraKeys,
    attributes,
    loggedInUserId
  );

  await Group.save({ ...group, ...groupPayload }, { data: { user } });

  const policyOperationResult = await bulkUpsertPoliciesForGroup(
    groupId,
    policies,
    loggedInUserId
  );
  const updatedGroup = await get(groupId, loggedInUserId);
  return { ...updatedGroup, policyOperationResult };
};
