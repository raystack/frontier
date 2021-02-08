import Hapi from '@hapi/hapi';
import * as Schema from './schema';
import * as Resource from './resource';

const iamConfig = {
  iam: {
    authorize: [
      {
        action: { baseName: 'iam' },
        resource: [{ requestKey: 'groupId', iamKey: 'group' }]
      }
    ]
  }
};

export const get = {
  description: 'fetch user and group mapping',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { groupId, userId } = request.params;
    return Resource.get(groupId, userId);
  }
};

export const post = {
  description: 'create group and user mapping',
  tags: ['api'],
  validate: {
    payload: Schema.payloadSchema
  },
  app: iamConfig,
  handler: async (request: Hapi.Request) => {
    const { payload } = request;
    const { groupId, userId } = request.params;
    const { id: loggedInUserId } = request.auth.credentials;
    return Resource.create(groupId, userId, loggedInUserId, payload);
  }
};

export const put = {
  description: 'update group and user mapping',
  tags: ['api'],
  validate: {
    payload: Schema.payloadSchema
  },
  app: iamConfig,
  handler: async (request: Hapi.Request) => {
    const { payload } = request;
    const { groupId, userId } = request.params;
    const { id: loggedInUserId } = request.auth.credentials;
    return Resource.update(groupId, userId, loggedInUserId, payload);
  }
};

export const remove = {
  description: 'delete group and user mapping',
  tags: ['api'],
  app: iamConfig,
  handler: async (request: Hapi.Request) => {
    const { groupId, userId } = request.params;
    const { id: loggedInUserId } = request.auth.credentials;
    return Resource.remove(groupId, userId, loggedInUserId);
  }
};
