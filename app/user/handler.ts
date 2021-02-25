import Hapi from '@hapi/hapi';
import * as Schema from './schema';
import * as Resource from './resource';

export const create = {
  description: 'create user',
  tags: ['api'],
  validate: {
    payload: Schema.createPayload
  },
  handler: async (request: Hapi.Request) => {
    return Resource.create(request.payload);
  }
};

export const list = {
  description: 'get list of users',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    return Resource.list(request.query);
  }
};

export const get = {
  description: 'get user by id',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    return Resource.get(request.params.userId);
  }
};
