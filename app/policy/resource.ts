import * as R from 'ramda';
import { createQueryBuilder } from 'typeorm';
import CasbinSingleton from '../../lib/casbin';
import { User } from '../../model/user';
import {
  toLikeQuery,
  parsePolicies,
  parsePoliciesWithSubject,
  extractResourceAction
} from './util';

type JSObj = Record<string, unknown>;
export interface PolicyOperation {
  operation: 'delete' | 'create';
  subject: JSObj;
  resource: JSObj;
  action: JSObj;
}
let casbinPolicies: any[] = [];

export const bulkOperation = async (
  policyOperations: PolicyOperation[] = [],
  subject: JSObj
) => {
  const promiseList = policyOperations.map(async ({ operation, ...policy }) => {
    // ? the subject who is performing the action should have iam.manage permission
    const hasAccess = await CasbinSingleton.enforcer?.enforceJson(
      subject,
      policy.resource,
      { action: 'iam.manage' }
    );
    if (!hasAccess) return false;

    switch (operation) {
      case 'create': {
        await CasbinSingleton.enforcer?.addJsonPolicy(
          policy.subject,
          policy.resource,
          policy.action
        );
        break;
      }
      case 'delete': {
        await CasbinSingleton.enforcer?.removeJsonPolicy(
          policy.subject,
          policy.resource,
          policy.action
        );
        break;
      }
      default: {
        break;
      }
    }
    return true;
  });

  const result = await Promise.all(promiseList);

  return policyOperations.map((policyOperation, index) => ({
    ...policyOperation,
    success: result[index]
  }));
};

export const getPoliciesBySubject = async (
  subject: JSObj,
  filters: JSObj = {}
) => {
  const { resource, action } = extractResourceAction(filters);

  const cursor = createQueryBuilder()
    .select('*')
    .from('casbin_rule', 'casbin_rule')
    .where('casbin_rule.ptype = :type', { type: 'p' })
    .andWhere('casbin_rule.v0 like :subject', {
      subject: toLikeQuery(subject)
    });

  if (!R.isEmpty(resource)) {
    cursor.andWhere('casbin_rule.v1 like :resource', {
      resource: toLikeQuery(resource || {})
    });
  }
  if (!R.isEmpty(action)) {
    cursor.andWhere('casbin_rule.v2 like :action', {
      action: toLikeQuery(action || {})
    });
  }
  const rawResult = await cursor.getRawMany();
  return parsePolicies(rawResult);
};

export const getSubjecListWithPolicies = async (
  subjectType: 'group' | 'user'
) => {
  const tableName = `${subjectType}s`;
  const columnName = 'id';
  const joinMatchStr = `${tableName}.${columnName}`;

  const rawResults = await createQueryBuilder()
    .select(
      `${joinMatchStr}, JSON_AGG(casbin_rule.*) FILTER (WHERE ptype = 'p') as policies, JSON_AGG(DISTINCT ${tableName}.*) AS ${subjectType}`
    )
    .from(tableName, tableName)
    .leftJoin(
      'casbin_rule',
      'casbin_rule',
      `casbin_rule.v0 LIKE '%"' || ${joinMatchStr} || '"%'`
    )
    .groupBy(joinMatchStr)
    .getRawMany();

  return parsePoliciesWithSubject(rawResults, subjectType);
};

export const getMapping = async (
  type: string,
  domain: string,
  first: JSObj = {},
  second: JSObj = {}
) => {
  const cursor = createQueryBuilder()
    .select('*')
    .from('casbin_rule', 'casbin_rule')
    .where('casbin_rule.ptype = :type', { type })
    .andWhere('casbin_rule.v2 = :domain', {
      domain
    });

  if (!R.isEmpty(first)) {
    cursor.andWhere('casbin_rule.v0 like :first', {
      first: toLikeQuery(first)
    });
  }

  if (!R.isEmpty(second)) {
    cursor.andWhere('casbin_rule.v1 like :second', {
      second: toLikeQuery(second)
    });
  }

  return cursor.getRawMany();
};

export const getGroupUserMapping = async (groupId: string, userId: string) => {
  const rawResult = await getMapping(
    'g',
    'subject',
    { user: userId },
    { group: groupId }
  );

  if (!R.isEmpty(rawResult)) {
    const { v0: userStr, v1: groupStr } = R.head(rawResult);
    return R.mergeAll([JSON.parse(userStr), JSON.parse(groupStr)]);
  }

  return null;
};

export const getAttributesForGroup = async (groupId: string) => {
  const rawResult = await getMapping('g2', 'resource', { group: groupId });
  return rawResult.map((rawObj) => JSON.parse(rawObj.v1));
};

export const getUsersOfGroupWithPolicies = async (
  groupId: string,
  policyFilter: JSObj = {}
) => {
  const { resource = {}, action = {} } = extractResourceAction(policyFilter);

  const POLICY_AGGREGATE = `JSON_AGG(casbin_rule.*) FILTER (WHERE casbin_rule.ptype = 'p') AS policies`;
  const GET_USER_DOC = `JSON_AGG(DISTINCT users.*) AS user`;

  // ! We need COUNT and SUM separate here because postgres doesn't support using COUNT twice on the same column
  const POLICY_COUNT = `COUNT(1) FILTER (WHERE casbin_rule.ptype = 'p')`;
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
    )
    .orWhere(
      `(casbin_rule.ptype = 'p' AND casbin_rule.v0 like :psubject AND casbin_rule.v1 like :resource AND casbin_rule.v2 like :action)`,
      {
        psubject: `%user%`,
        resource: toLikeQuery({ group: groupId, ...resource }),
        action: toLikeQuery(action)
      }
    )
    .groupBy('users.id');

  if (!R.isEmpty(policyFilter)) {
    cursor.having(`${POLICY_COUNT} > 0`);
  }
  const rawResult = await cursor.andHaving(`${MAPPING_COUNT} > 0`).getRawMany();

  return parsePoliciesWithSubject(rawResult, 'user');
};

export const setPolicies = (policies: [] = []) => {
  casbinPolicies = policies;
};

export const getPolicies = () => {
  return casbinPolicies;
};
