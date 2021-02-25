import * as R from 'ramda';
import { createQueryBuilder } from 'typeorm';
import { Group } from '../../../model/group';
import { toLikeQuery } from '../../policy/util';
import { parseGroupListResult } from './util';
import CasbinSingleton from '../../../lib/casbin/index';

type JSObj = Record<string, unknown>;

const getImplicitGroups = async (userId: string, filters: JSObj = {}) => {
  const { action, ...attributes } = filters;

  // fetch all groups filtered on attributes
  const GET_GROUP_DOC = `JSON_AGG(DISTINCT groups.*) AS group_arr`;
  const ATTRIBUTES_AGGREGATE = `JSON_AGG(casbin_rule.*) FILTER (WHERE casbin_rule.ptype = 'g2') AS raw_attributes`;

  const JOIN_WITH_RESOURCE_ATTRIBUTES_MAPPING = `casbin_rule.ptype = 'g2' AND casbin_rule.v0 like '%"' || groups.id || '"%'`;

  const cursor = createQueryBuilder()
    .select(`groups.id, ${GET_GROUP_DOC}, ${ATTRIBUTES_AGGREGATE}`)
    .from(Group, 'groups')
    .leftJoin(
      'casbin_rule',
      'casbin_rule',
      `${JOIN_WITH_RESOURCE_ATTRIBUTES_MAPPING}`
    )
    .groupBy('groups.id');

  if (!R.isEmpty(attributes)) {
    const FILTER_BY_RESOURCE_ATTRIBUTES = `SUM(CASE WHEN casbin_rule.ptype = 'g2' AND casbin_rule.v1 like :attributes THEN 1 ELSE 0 END) > 0`;
    cursor.having(FILTER_BY_RESOURCE_ATTRIBUTES, {
      attributes: toLikeQuery(attributes)
    });
  }

  const rawResult = await cursor.getRawMany();
  const allGroups = parseGroupListResult(rawResult);

  // batchEnforce to check whether {user: userId}, {group: group.id}, {action}
  const groupPoliciesToEnforce = allGroups.map((group: any) => ({
    subject: { user: userId },
    resource: { group: group.id },
    action: { action }
  }));

  const batchEnforceResult =
    (await CasbinSingleton.enforcer?.batchEnforceJson(
      groupPoliciesToEnforce
    )) || [];

  return allGroups.filter((group, index) => batchEnforceResult[index]);
};

const getExplicitGroups = async (userId: string, attributes: JSObj = {}) => {
  const userQuery = JSON.stringify({ user: userId });

  const GET_GROUP_DOC = `JSON_AGG(DISTINCT groups.*) AS group_arr`;
  const ATTRIBUTES_AGGREGATE = `JSON_AGG(casbin_rule.*) FILTER (WHERE casbin_rule.ptype = 'g2') AS raw_attributes`;
  const AGGREGATE_MEMBER_POLICIES = `JSON_AGG(casbin_rule.*) FILTER (WHERE casbin_rule.ptype = 'p' AND casbin_rule.v0 = '${userQuery}') AS raw_policies`;

  const JOIN_WITH_RESOURCE_ATTRIBUTES_MAPPING = `casbin_rule.ptype = 'g2' AND casbin_rule.v0 like '%"' || groups.id || '"%'`;
  const JOIN_WITH_USER_MAPPING = `casbin_rule.ptype = 'g' AND casbin_rule.v0 = '${userQuery}' AND casbin_rule.v1 like '%"' || groups.id || '"%'`;
  const JOIN_WITH_POLICIES = `casbin_rule.ptype = 'p' AND casbin_rule.v0 = '${userQuery}'  AND casbin_rule.v1 like '%"' || groups.id || '"%'`;

  const cursor = createQueryBuilder()
    .select(
      `groups.id, ${GET_GROUP_DOC}, ${ATTRIBUTES_AGGREGATE}, ${AGGREGATE_MEMBER_POLICIES}`
    )
    .from(Group, 'groups')
    .leftJoin(
      'casbin_rule',
      'casbin_rule',
      `(${JOIN_WITH_RESOURCE_ATTRIBUTES_MAPPING}) OR (${JOIN_WITH_USER_MAPPING}) OR (${JOIN_WITH_POLICIES})`
    )
    .having(`COUNT(*) FILTER (WHERE ${JOIN_WITH_USER_MAPPING}) > 0`)
    .groupBy('groups.id');

  if (!R.isEmpty(attributes)) {
    const FILTER_BY_RESOURCE_ATTRIBUTES = `SUM(CASE WHEN casbin_rule.ptype = 'g2' AND casbin_rule.v1 like :attributes THEN 1 ELSE 0 END) > 0`;
    cursor.andHaving(FILTER_BY_RESOURCE_ATTRIBUTES, {
      attributes: toLikeQuery(attributes)
    });
  }

  const rawResult = await cursor.getRawMany();
  return parseGroupListResult(rawResult);
};

export const list = async (userId: string, filters: JSObj = {}) => {
  const { action } = filters;
  if (!R.isNil(action)) {
    return getImplicitGroups(userId, filters);
  }
  return getExplicitGroups(userId, filters);
};
