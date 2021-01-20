import Hapi from '@hapi/hapi';
import * as Schema from './schema';
import * as Resource from './resource';

export const get = {
  description: 'get group by id',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    return Resource.get(request.params.id);
  }
};

export const create = {
  description: 'create group',
  tags: ['api'],
  validate: {
    payload: Schema.createPayload
  },
  handler: async (request: Hapi.Request) => {
    return Resource.create(request.payload);
  }
};

export const update = {
  description: 'update group by id',
  tags: ['api'],
  validate: {
    payload: Schema.updatePayload
  },
  handler: async (request: Hapi.Request) => {
    return Resource.update(request.params.id, request.payload);
  }
};
