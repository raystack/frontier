import * as R from 'ramda';

export const extractRoleTagFilter = R.pathOr(null, [
  'fields',
  'policies',
  '$filter',
  'role',
  'tags'
]);
