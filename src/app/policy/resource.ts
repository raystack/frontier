import * as R from 'ramda';
import { createQueryBuilder } from 'typeorm';
import CasbinSingleton from '../../lib/casbin';
import { convertJSONToStringInOrder } from '../../lib/casbin/JsonFilteredEnforcer';
import { User } from '../../model/user';
import { extractRoleTagFilter } from '../../utils/queryParams';
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

export const bulkOperation = async (
  policyOperations: PolicyOperation[] = [],
  subject: JSObj
) => {
  const user = await User.findOne({
    where: {
      id: subject.user
    }
  });
  const promiseList = policyOperations.map(async ({ operation, ...policy }) => {
    // ? the subject who is performing the action should have iam.manage permission
    const hasAccess = await CasbinSingleton.enforcer?.enforceJson(
      subject,
      policy.resource,
      { action: 'iam.manage' }
    );
    if (!hasAccess) return false;

    const options: JSObj = { created_by: user };
    switch (operation) {
      case 'create': {
        await CasbinSingleton.enforcer?.addJsonPolicy(
          policy.subject,
          policy.resource,
          policy.action,
          options
        );
        break;
      }
      case 'delete': {
        await CasbinSingleton.enforcer?.removeJsonPolicy(
          policy.subject,
          policy.resource,
          policy.action,
          { created_by: user }
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
  const resourceActionFilter = R.omit(['fields'], filters);
  const { resource, action } = extractResourceAction(resourceActionFilter);
  const roleTag = extractRoleTagFilter(filters);

  const cursor = createQueryBuilder()
    .select('casbin_rule.*')
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

  if (!R.isNil(roleTag)) {
    cursor
      .leftJoin(
        'roles',
        'roles',
        `casbin_rule.v2 like '%"' || roles.id || '"%'`
      )
      .andWhere(`:roleTag = ANY(roles.tags)`, {
        roleTag
      });
  }
  const rawResult = await cursor.getRawMany();
  return parsePolicies(rawResult);
};

export const getSubjecListWithPolicies = async (
  subjectType: 'group' | 'user',
  roleTags: string[] = []
) => {
  const tableName = `${subjectType}s`;
  const columnName = 'id';
  const joinMatchStr = `${tableName}.${columnName}`;
  const roleTagQuery =
    roleTags.length > 1
      ? `AND "roles"."tags" @> '{${roleTags.join(',')}}'`
      : '';
  const policyAggregateStr = `JSON_AGG(casbin_rule.*) FILTER (WHERE ptype = 'p' ${roleTagQuery}) as policies`;
  const cursor = createQueryBuilder()
    .select(
      `${joinMatchStr}, ${policyAggregateStr}, JSON_AGG(DISTINCT ${tableName}.*) AS ${subjectType}`
    )
    .from(tableName, tableName)
    .leftJoin(
      'casbin_rule',
      'casbin_rule',
      `casbin_rule.v0 = '{"${subjectType}":"' || ${joinMatchStr} || '"}'`
    )
    .leftJoin('roles', 'roles', `casbin_rule.v2 like '%"' || roles.id || '"%'`);
  const rawResults = await cursor.groupBy(joinMatchStr).getRawMany();
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

export const getResourceAttributeMappingsByResources = async (
  resources: JSObj[] = []
) => {
  const stringifiedResources = resources.map((res: JSObj) =>
    convertJSONToStringInOrder(res)
  );

  const rawResult = await createQueryBuilder()
    .select('*')
    .from('casbin_rule', 'casbin_rule')
    .where('casbin_rule.ptype = :type', { type: 'g2' })
    .andWhere('casbin_rule.v2 = :domain', {
      domain: 'resource'
    })
    .andWhere('casbin_rule.v0 in (:...resources)', {
      resources: stringifiedResources
    })
    .getRawMany();

  const parsedResults = rawResult.map((res) => ({
    resource: JSON.parse(R.propOr('{}', 'v0', res)),
    attributes: JSON.parse(R.propOr('{}', 'v1', res))
  }));

  return parsedResults;
};
