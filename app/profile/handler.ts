import Hapi from '@hapi/hapi';
import * as Schema from './schema';
import * as Resource from './resource';

export const get = {
  description: "get current user's profile",
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { id }: any = request.auth?.credentials;
    return Resource.getUserById(id);
  }
};

export const update = {
  description: 'update current user',
  tags: ['api'],
  validate: {
    payload: Schema.updatePayload
  },
  handler: async (request: Hapi.Request) => {
    const { id }: any = request.auth?.credentials;
    return Resource.updateUserById(id, request.payload);
  }
};
