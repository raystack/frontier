import Boom from '@hapi/boom';
import * as R from 'ramda';
import { createQueryBuilder } from 'typeorm';
import CasbinSingleton from '../../../lib/casbin';
import {
  bulkOperation,
  getPoliciesBySubject,
  getGroupUserMapping
} from '../../policy/resource';
import { toLikeQuery, parsePoliciesWithSubject } from '../../policy/util';
import { User } from '../../../model/user';
import { extractRoleTagFilter } from '../../../utils/queryParams';

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
  loggedInUserId: string
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

  const policies = await getPoliciesBySubject(userObj, groupObj);

  if (!R.isEmpty(policies)) {
    const promiseList = policies.map(async (policy) => {
      await CasbinSingleton.enforcer?.removeJsonPolicy(
        policy.subject,
        policy.resource,
        policy.action,
        options
      );
    });
    await Promise.all(promiseList);
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

export const list = async (
  groupId: string,
  options: Record<string, unknown> = {}
) => {
  const roleTag = extractRoleTagFilter(options);

  const POLICY_AGGREGATE = `JSON_AGG(casbin_rule.*) FILTER (WHERE casbin_rule.ptype = 'p') AS policies`;
  const GET_USER_DOC = `JSON_AGG(DISTINCT users.*) AS user`;

  const MAPPING_COUNT = `SUM(CASE WHEN casbin_rule.ptype = 'g' THEN 1 ELSE 0 END)`;

  // ? Join users table with casbin_rule table and groupBy user.id
  // ? Aggregate all the policies
  // ? Remove users that don't have a 'g' record with the groupId
  const cursor = createQueryBuilder()
    .select(`users.id, ${POLICY_AGGREGATE}, ${GET_USER_DOC}`)
    .from(User, 'users')
    .leftJoin(
      'casbin_rule',
      'casbin_rule',
      `casbin_rule.v0 like '%"' || users.id || '"%'`
    )
    .where(
      `(casbin_rule.ptype = 'g' AND casbin_rule.v0 like :gsubject AND casbin_rule.v1 like :group)`,
      {
        gsubject: `%user%`,
        group: toLikeQuery({ group: groupId })
      }
    );

  if (!R.isNil(roleTag)) {
    cursor
      .leftJoin(
        'roles',
        'roles',
        `casbin_rule.v2 like '%"' || roles.id || '"%'`
      )
      .orWhere(
        `(casbin_rule.ptype = 'p' AND casbin_rule.v0 like :psubject AND casbin_rule.v1 like :resource AND :roleTag = ANY(roles.tags))`,
        {
          psubject: `%user%`,
          resource: toLikeQuery({ group: groupId }),
          roleTag
        }
      );
  } else {
    cursor.orWhere(
      `(casbin_rule.ptype = 'p' AND casbin_rule.v0 like :psubject AND casbin_rule.v1 like :resource)`,
      {
        psubject: `%user%`,
        resource: toLikeQuery({ group: groupId })
      }
    );
  }

  const rawResult = await cursor
    .groupBy('users.id')
    .andHaving(`${MAPPING_COUNT} > 0`)
    .getRawMany();

  return parsePoliciesWithSubject(rawResult, 'user');
};
