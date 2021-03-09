import Boom from '@hapi/boom';
import * as R from 'ramda';
import CasbinSingleton from '../../../lib/casbin';
import {
  bulkOperation,
  getPoliciesBySubject,
  getGroupUserMapping,
  getUsersOfGroupWithPolicies,
  PolicyOperation
} from '../../policy/resource';
import { User } from '../../../model/user';

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
  const user = await User.findOne({
    where: {
      id: loggedInUserId
    }
  });
  const options = { created_by: user };
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  await CasbinSingleton.enforcer?.addSubjectGroupingJsonPolicy(
    subject,
    group,
    options
  );
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
  groupId: string,
  userId: string,
  loggedInUserId: any
) => {
  const userObj = { user: userId };
  const groupObj = { group: groupId };
  const user = await User.findOne({
    where: {
      id: loggedInUserId
    }
  });
  const options = { created_by: user };
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  await CasbinSingleton.enforcer?.removeSubjectGroupingJsonPolicy(
    userObj,
    groupObj,
    options
  );

  const policies = (await getPoliciesBySubject(userObj, groupObj)).map(
    R.assoc('operation', 'delete')
  );
  if (!R.isEmpty(policies)) {
    await bulkOperation(<PolicyOperation[]>policies, { user: loggedInUserId });
  }

  return true;
};

export const get = async (groupId: string, userId: string) => {
  const userObj = { user: userId };
  const groupObj = { group: groupId };

  const groupUserMapping = await getGroupUserMapping(groupId, userId);
  if (R.isNil(groupUserMapping)) {
    return Boom.notFound('user not found in group');
  }

  const user = await User.findOne(userId);
  const policies = await getPoliciesBySubject(userObj, groupObj);

  return { ...user, policies };
};

export const list = getUsersOfGroupWithPolicies;
