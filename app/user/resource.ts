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

// /api/users?entity=gojek&privacy=public&action=firehose.read
export const getListWithFilters = async (policyFilters: JSObj) => {
  // ? 1) Get all users with all policies
  const allUsersWithAllPolicies = await getSubjecListWithPolicies('user');

  // 3) fetch all groups with the matching attributes
  const rawGroupResult = await createQueryBuilder()
    .select('*')
    .from('casbin_rule', 'casbin_rule')
    .where('casbin_rule.ptype = :type', { type: 'g2' })
    .andWhere('casbin_rule.v1 = :filter', {
      filter: JSON.stringify(policyFilters)
    })
    .getRawMany();

  const groups = rawGroupResult.map((res) => res.v0);

  if (groups.length === 0) return [];

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

  // 5) only return users that match the users<->groups mapping
  const usersWithAccesss = allUsersWithAllPolicies.filter((user: any) => {
    return userMap[user.id];
  });

  if (R.isEmpty(policyFilters)) return usersWithAccesss;

  // 4) fetch policies of the selected users, filtered by either of the policyFilter
  const usersWithFilteredPolicies = usersWithAccesss.map((user: any) => {
    const { policies = [] } = user;
    const filteredPolicies = policies.filter((policy: JSObj) =>
      isJSONSubset(
        JSON.stringify(policyFilters),
        JSON.stringify(policy.resource)
      )
    );
    return R.assoc('policies', filteredPolicies, user);
  });

  return usersWithFilteredPolicies;
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
