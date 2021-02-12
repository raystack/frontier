import * as R from 'ramda';
import { User } from '../../model/user';
import { extractResourceAction } from '../policy/util';
import { getSubjecListWithPolicies } from '../policy/resource';
import CasbinSingleton from '../../lib/casbin';
import { isJSONSubset } from '../../lib/casbin/JsonRoleManager';

type JSObj = Record<string, unknown>;

export const create = async (payload: any) => {
  return await User.save(payload);
};

// /api/users?entity=gojek&privacy=public&action=firehose.read
export const getListWithFilters = async (policyFilters: JSObj) => {
  // ? 1) Get all users with all policies
  const allUsersWithAllPolicies = await getSubjecListWithPolicies('user');

  // 2) extract resource and action from policyFilters
  const { resource = {}, action = {} } = extractResourceAction(policyFilters);

  // 3) run each user through casbin enforcer based on the specifed params
  const enforcedUsers = await Promise.all(
    allUsersWithAllPolicies.map(async (user: any) => {
      const hasAccess = await CasbinSingleton?.enforcer?.enforceJson(
        { user: user.id },
        resource,
        action
      );
      return R.assoc('hasAccess', hasAccess, user);
    })
  );

  const usersWithAccesss = enforcedUsers.filter((user: any) => user.hasAccess);

  if (R.isEmpty(resource)) return usersWithAccesss;

  // 4) fetch policies of the selected users, filtered by either of the policyFilter
  const usersWithFilteredPolicies = usersWithAccesss.map((user: any) => {
    const { policies = [] } = user;
    const filteredPolicies = policies.filter((policy: JSObj) =>
      isJSONSubset(JSON.stringify(resource), JSON.stringify(policy.resource))
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
