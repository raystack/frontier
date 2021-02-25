import * as R from 'ramda';
import { toCamelCase, parsePolicies } from '../../policy/util';

type JSObj = Record<string, undefined>;

type GroupRawResult = {
  id: string;
  raw_attributes?: JSObj[];
  group_arr: JSObj[];
  raw_policies?: JSObj[];
};

const parseGroupResult = (group: GroupRawResult) => {
  const {
    raw_attributes: rawAttributes,
    group_arr: groupArr,
    raw_policies: rawPolicies
  } = group;

  const attributes = rawAttributes?.map(
    R.pipe(R.propOr('{}', 'v1'), JSON.parse)
  );
  const parsedGroup = R.head(groupArr);

  return {
    ...toCamelCase(parsedGroup || {}),
    policies: parsePolicies(rawPolicies || []),
    attributes
  };
};

export const parseGroupListResult = R.map(parseGroupResult);
