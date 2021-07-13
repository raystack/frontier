import * as R from 'ramda';
import * as _ from 'lodash';
import { convertJSONToStringInOrder } from '../../lib/casbin/JsonEnforcer';

type JSObj = Record<string, unknown>;

export const toCamelCase = (data: JSObj) =>
  _.mapKeys(data, (value, key) => _.camelCase(key));

export const toLikeQuery = (json: JSObj = {}) =>
  `%${convertJSONToStringInOrder(json)
    .replace('{', '')
    .replace('}', '')
    .replace(/:/g, ':%')
    .replace(/,/g, '%,%')}%`;

export const parsePolicies = (rawPolicies: JSObj[]) => {
  return rawPolicies.map(({ v0, v1, v2 }) => ({
    subject: JSON.parse(<string>v0),
    resource: JSON.parse(<string>v1),
    action: JSON.parse(<string>v2)
  }));
};

export const extractResourceAction = (data: JSObj = {}) => {
  const ACTION_KEYS = ['action', 'role'];
  const action = R.pick(ACTION_KEYS, data);
  const resource = R.omit(ACTION_KEYS, data);
  return { resource, action };
};

export const parsePoliciesWithSubject = (
  rawList: JSObj[] = [],
  subjectType: string
) => {
  return rawList.map((rawObj: JSObj) => {
    // ? group is a list here because of json_agg function, but the length of the array will always be 1
    const subjectAgg = <Array<JSObj>>R.propOr([{}], subjectType, rawObj);
    const policies = <Array<JSObj>>R.propOr([], 'policies', rawObj);
    const subjectWithCamelKeys = toCamelCase(R.head(subjectAgg) || {});
    return { ...subjectWithCamelKeys, policies: parsePolicies(policies) };
  });
};
