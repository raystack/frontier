import Wreck from '@hapi/wreck';
import * as R from 'ramda';
import { User } from '../../model/user';
import { Group } from '../../model/group';
import { getUsernameFromIAPHeader } from '../iap/utils';

const pipeWithPromise = R.pipeWith((fun, previousResult) =>
  previousResult && previousResult.then
    ? previousResult.then(fun)
    : fun(previousResult)
);

export const checkAndAppendUserData = async (body: any) => {
  if (R.has('user_id', body)) {
    const user = await User.findOne(body.user_id);
    return R.assocPath(['user'], user, body);
  }

  return body;
};

export const checkAndAppendGroupData = async (body: any) => {
  if (R.has('group_id', body)) {
    const group = await Group.findOne(body.group_id);
    return R.assocPath(['group'], group, body);
  }

  return body;
};

export const checkAndAppendCreatedBy = (modifiedBody: any) => {
  if (R.hasPath(['user', 'username'], modifiedBody)) {
    return R.assocPath(
      ['created_by'],
      R.path(['user', 'username'], modifiedBody),
      modifiedBody
    );
  }
  return modifiedBody;
};

export const checkAndAppendTeam = (modifiedBody: any) => {
  if (R.hasPath(['group', 'name'], modifiedBody)) {
    return R.assocPath(
      ['team'],
      R.path(['group', 'name'], modifiedBody),
      modifiedBody
    );
  }
  return modifiedBody;
};

// TODO: Remove this once odin and hawkeye handle this
export const checkAndAppendDataForOdin = async (body: any) => {
  return R.pipe(checkAndAppendCreatedBy, checkAndAppendTeam)(body);
};

export const appendProxyPayload = async (options: any) => {
  if (options.payload) {
    const body = await Wreck.read(options.payload, {
      json: 'force',
      gunzip: true
    });

    const modifiedPayload = await pipeWithPromise([
      checkAndAppendUserData,
      checkAndAppendGroupData,
      checkAndAppendDataForOdin
    ])(body);

    return R.assocPath(['payload'], modifiedPayload, options);
  }

  return options;
};

export const appendHeaders = async (options: any) => {
  const googleIAPHeader =
    R.pathOr('', ['headers', 'x-goog-authenticated-user-email'], options) ||
    R.pathOr('', ['headers', 'X-Goog-Authenticated-User-Email'], options);

  const username = getUsernameFromIAPHeader(googleIAPHeader);
  if (!R.isEmpty(username)) {
    return R.assocPath(['headers', 'username'], username, options);
  }
  return options;
};

// TODO: Append username in headers as well
export default async (options: any) => {
  const optionsWithPayload = await appendProxyPayload(options);
  return appendHeaders(optionsWithPayload);
};
