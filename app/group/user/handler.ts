import Hapi from '@hapi/hapi';
import Boom from '@hapi/boom';
import * as Schema from './schema';
import * as Resource from './resource';
import CasbinSingleton from '../../../lib/casbin';

const hasAccess = (request: Hapi.Request) => {
  if (request.method === 'get') {
    return true; // get api is open for all
  }

  const { id: userId } = request.auth.credentials;
  const { groupIdentifier } = request.params;
  let action = 'iam.create';
  if (request.method === 'put') {
    action = 'iam.manage';
  } else if (request.method === 'delete') {
    action = '';
  }
  return CasbinSingleton.enforcer?.enforceJson(
    { user: userId },
    {
      group: groupIdentifier
    },
    {
      action
    }
  );
};

export const operation = {
  description: 'group and user mapping for create/update/delete resource',
  tags: ['api'],
  validate: {
    payload: Schema.payloadSchema
  },
  handler: async (request: Hapi.Request) => {
    const { payload } = request;
    const { groupId, userId } = request.params;
    const { id: loggedInUserId } = request.auth.credentials;

    if (!hasAccess(request)) {
      return Boom.unauthorized(
        "Sorry you don't have permission to be perform this action"
      );
    }

    switch (request.method) {
      case 'get':
        return Resource.get(groupId, userId);
      case 'post':
        return Resource.create(groupId, userId, loggedInUserId, payload);
      case 'put':
        return Resource.update(groupId, userId, loggedInUserId, payload);
      case 'delete':
        return Resource.remove(groupId, userId, loggedInUserId, payload);
      default:
        throw new Error('Requested method not supported');
    }
  }
};
