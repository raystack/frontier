import * as R from 'ramda';
import { createQueryBuilder } from 'typeorm';
import { User } from '../../model/user';
import { getSubjecListWithPolicies } from '../policy/resource';
import { isJSONSubset } from '../../lib/casbin/JsonRoleManager';
import getUniqName, { validateUniqName } from '../../lib/getUniqName';

type JSObj = Record<string, unknown>;

const getValidUsername = async (payload: any) => {
  let username = payload?.username;
  if (payload?.displayname && !username) {
    username = await getUniqName(payload?.displayname, 'username', User);
  }
  validateUniqName(username);
  return username;
};

export const create = async (payload: any) => {
  const username = await getValidUsername(payload);
  return await User.save({ ...payload, username });
};

// /api/users?entity=gojek
export const getListWithFilters = async (query: JSObj) => {
  // ? 1) Get all users with all policies

  const roleTagFilter = R.pathOr(null, [
    'fields',
    'policies',
    '$filter',
    'role',
    'tag'
  ])(query);

  const policyFilters = R.omit(['fields'], query);

  const allUsersWithAllPolicies = await getSubjecListWithPolicies(
    'user',
    roleTagFilter
  );

  // 3) fetch all groups with the matching attributes
  const rawGroupResultCursor = createQueryBuilder()
    .select('*')
    .from('casbin_rule', 'casbin_rule');

  if (!R.isEmpty(policyFilters)) {
    rawGroupResultCursor
      .where('casbin_rule.ptype = :type', { type: 'g2' })
      .andWhere('casbin_rule.v1 = :filter', {
        filter: JSON.stringify(policyFilters)
      });
  }

  const rawGroupResult = await rawGroupResultCursor.getRawMany();

  const groups = rawGroupResult.map((res) => res.v0);

  if (R.isEmpty(policyFilters) && R.isNil(roleTagFilter)) return [];

  // 4) fetch all groups_users record based on above groups
  const rawUserGroupResult = await createQueryBuilder()
    .select('*')
    .from('casbin_rule', 'casbin_rule')
    .where('casbin_rule.ptype = :type', { type: 'g' })
    .andWhere('casbin_rule.v1 in (:...groups)', {
      groups
    })
    .getRawMany();

  const userMap = rawUserGroupResult.reduce((uMap, rGUser) => {
    const userDoc = JSON.parse(rGUser.v0);
    // eslint-disable-next-line no-param-reassign
    uMap[userDoc.user] = 1;
    return uMap;
  }, {});

  // 5) only return users that match the users<->groups mapping or with matching policy
  const usersWithAccesss = allUsersWithAllPolicies.reduce(
    (result: any, user: any) => {
      const { policies = [] } = user;
      const filteredPolicies = policies.filter((policy: JSObj) =>
        !R.isEmpty(policyFilters)
          ? isJSONSubset(
              JSON.stringify(policyFilters),
              JSON.stringify(policy.resource)
            )
          : policies
      );

      const userHasAccess = userMap[user.id] || !R.isEmpty(filteredPolicies);
      if (userHasAccess) {
        const userWithPolicy = R.assoc('policies', filteredPolicies, user);
        result.push(userWithPolicy);
      }
      return result;
    },
    []
  );

  return usersWithAccesss;
};

export const list = async (policyFilters: JSObj = {}) => {
  if (R.isEmpty(policyFilters)) {
    return User.find();
  }

  return getListWithFilters(policyFilters);
};

export const get = async (id: string) => {
  return User.findOne({
    where: {
      id
    }
  });
};
