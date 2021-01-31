import * as R from 'ramda';
import { createQueryBuilder } from 'typeorm';
import CasbinSingleton from '../../lib/casbin';
import { convertJSONToStringInOrder } from '../../lib/casbin/JsonEnforcer';

type JSObj = Record<string, unknown>;
export interface PolicyOperation {
  operation: 'delete' | 'create';
  subject: JSObj;
  resource: JSObj;
  action: JSObj;
}

const toLikeQuery = (json: JSObj = {}) =>
  `%${convertJSONToStringInOrder(json)
    .replace('{', '')
    .replace('}', '')
    .replace(/:/g, ':%')
    .replace(/,/g, '%,%')}%`;

const parsePolicies = (rawPolicies: JSObj[]) => {
  return rawPolicies.map(({ v0, v1, v2 }) => ({
    subject: JSON.parse(<string>v0),
    resource: JSON.parse(<string>v1),
    action: JSON.parse(<string>v2)
  }));
};

const extractResourceAction = (data: JSObj = {}) => {
  const ACTION_KEYS = ['action', 'role'];
  const action = R.pick(ACTION_KEYS, data);
  const resource = R.omit(ACTION_KEYS, data);
  return { resource, action };
};

const parsePoliciesWithSubject = (rawList: JSObj[] = []) => {
  return rawList.map((rawObj: JSObj) => {
    // ? group is a list here because of json_agg function, but the length of the array will always be 1
    const groupList = <Array<JSObj>>R.propOr([{}], 'group', rawObj);
    const policies = <Array<JSObj>>R.propOr([], 'policies', rawObj);
    return { ...groupList.pop(), policies: parsePolicies(policies) };
  });
};

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

const subjectTypeToTableNameMap = {
  group: 'groups',
  username: 'users'
};

const subjectTypeToTableColumnNameMap = {
  group: 'name',
  username: 'username'
};

export const getSubjecListWithPolicies = async (
  subjectType: string,
  filters: JSObj = {}
) => {
  const { resource, action } = extractResourceAction(filters);

  const tableName = <string>R.path([subjectType], subjectTypeToTableNameMap);
  const columnName = <string>(
    R.path([subjectType], subjectTypeToTableColumnNameMap)
  );
  const joinMatchStr = `${tableName}.${columnName}`;

  const cursor = createQueryBuilder()
    .select(
      `${joinMatchStr}, json_agg(casbin_rule.*) as policies, json_agg(DISTINCT ${tableName}.*) as ${subjectType}`
    )
    .from(tableName, tableName)
    .where('casbin_rule.ptype = :type', { type: 'p' })
    .andWhere('casbin_rule.v0 like :subject', {
      subject: `%"${subjectType}"%`
    })
    .leftJoin(
      'casbin_rule',
      'casbin_rule',
      `casbin_rule.v0 like '%"' || ${joinMatchStr} || '"%'`
    )
    .groupBy(joinMatchStr);

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

  const rawResults = await cursor.getRawMany();
  return parsePoliciesWithSubject(rawResults);
};
