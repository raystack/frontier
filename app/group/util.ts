import * as R from 'ramda';
import { toCamelCase, parsePolicies } from '../policy/util';

type JSObj = Record<string, undefined>;

type GroupRawResult = {
  is_member: string;
  member_count: string;
  raw_attributes?: JSObj[];
  id: string;
  displayName: string;
  metadata: JSObj;
  group_arr: JSObj[];
  raw_user_policies?: JSObj[];
};

const parseGroupResult = (group: GroupRawResult) => {
  const {
    raw_attributes: rawAttributes,
    is_member: isMember,
    member_count: memberCount,
    group_arr: groupArr,
    raw_user_policies: rawUserPolicies
  } = group;

  const attributes = rawAttributes?.map(
    R.pipe(R.propOr('{}', 'v1'), JSON.parse)
  );
  const parsedGroup = R.head(groupArr);

  return {
    ...toCamelCase(parsedGroup || {}),
    isMember,
    userPolicies: parsePolicies(rawUserPolicies || []),
    memberCount: +memberCount,
    attributes
  };
};

export const parseGroupListResult = R.map(parseGroupResult);
