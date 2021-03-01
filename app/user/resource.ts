import * as R from 'ramda';
import { User } from '../../model/user';
import { extractResourceAction } from '../policy/util';
import { getSubjecListWithPolicies } from '../policy/resource';
import CasbinSingleton from '../../lib/casbin';
import { isJSONSubset } from '../../lib/casbin/JsonRoleManager';
import getUniqName from '../../lib/getUniqName';

type JSObj = Record<string, unknown>;

export const create = async (payload: any) => {
  const basename = payload?.username || payload?.displayname;
  const username = await getUniqName(basename, 'username', User);
  return await User.save({ ...payload, username });
};

// /api/users?entity=gojek&privacy=public&action=firehose.read
export const getListWithFilters = async (policyFilters: JSObj) => {
  // ? 1) Get all users with all policies
  const allUsersWithAllPolicies = await getSubjecListWithPolicies('user');

  // 2) extract resource and action from policyFilters
  const { resource = {}, action = {} } = extractResourceAction(policyFilters);

  // 3) run each user through casbin enforcer based on the specifed params
  const policiesToBatchEnforce = allUsersWithAllPolicies.map((user: any) => ({
    subject: { user: user.id },
    resource,
    action
  }));
  const batchEnforceResults = await CasbinSingleton?.enforcer?.batchEnforceJson(
    policiesToBatchEnforce
  );

  const usersWithAccesss = allUsersWithAllPolicies.filter(
    (user: any, index: number) =>
      batchEnforceResults && batchEnforceResults[index]
  );

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

export const get = async (id: string) => {
  return User.findOne({
    where: {
      id
    }
  });
};
