import CasbinSingleton from '../../../lib/casbin';
import { bulkOperation } from '../../policy/resource';

export const create = async (
  groupId: string,
  userId: string,
  loggedInUserId: any,
  payload: any
) => {
  const { policies = [] } = payload;
  const subject = {
    user: userId
  };
  const group = {
    group: groupId
  };

  await CasbinSingleton.enforcer?.addSubjectGroupingJsonPolicy(subject, group);
  return await bulkOperation(policies, { user: loggedInUserId });
};

export const update = async (
  groupId: string,
  userId: string,
  loggedInUserId: any,
  payload: any
) => {
  const { policies = [] } = payload;
  const subject = {
    user: loggedInUserId
  };

  return await bulkOperation(policies, subject);
};

export const remove = async (
  groupIdentifier: string,
  userIdentifier: string,
  loggedInUserId: any,
  payload: any
) => {
  const { policies = [] } = payload;
  console.log('remove => ', groupIdentifier, userIdentifier, policies);
};

export const get = async (groupIdentifier: string, userIdentifier: string) => {
  console.log('get => ', groupIdentifier, userIdentifier);
};
