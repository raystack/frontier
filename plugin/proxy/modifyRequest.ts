import * as R from 'ramda';
import Hapi from '@hapi/hapi';
import Boom from '@hapi/boom';
import { Group } from '../../model/group';

export const checkAndAppendUserData = (
  request: Hapi.Request,
  parsedPayload: any
) => {
  const user = request?.auth?.credentials || {};
  if (user?.username) {
    return R.pipe(
      R.assocPath(['user'], user),
      R.assocPath(['created_by'], user.username)
    )(parsedPayload);
  }

  return parsedPayload;
};

export const checkAndAppendGroupData = async (parsedPayload: any) => {
  if (R.has('group_id', parsedPayload)) {
    const group = await Group.findOne(R.propOr('', 'group_id', parsedPayload));

    if (!group) throw Boom.notFound('Group not found');

    return R.pipe(
      R.assocPath(['group'], group),
      R.assocPath(['team'], group?.groupname)
    )(parsedPayload);
  }

  return parsedPayload;
};

export const getModifiedPayload = async (request: Hapi.Request) => {
  const { payload } = request;
  try {
    const parsedPayload = JSON.parse(payload.toString());
    const payloadWithUser = checkAndAppendUserData(request, parsedPayload);
    const payloadWithGroupAndUser = await checkAndAppendGroupData(
      payloadWithUser
    );
    return payloadWithGroupAndUser;
    // eslint-disable-next-line no-empty
  } catch (e) {}

  return payload;
};

export const getModifiedHeaders = (request: Hapi.Request) => {
  const { username } = request?.auth?.credentials || {};
  if (username) {
    return R.assocPath(['username'], username, request.headers);
  }
  return request.headers;
};

export default async (request: any, h: Hapi.ResponseToolkit) => {
  if (request.payload) {
    request.payload = await getModifiedPayload(request);
  }
  request.headers = getModifiedHeaders(request);

  return h.continue;
};
