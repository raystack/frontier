import Hapi from '@hapi/hapi';
import * as R from 'ramda';
import * as Resource from './resource';

export const get = {
  description: "get current user's profile",
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { email }: any = request.auth?.credentials;
    return Resource.getUserByEmail(email);
  }
};

export const update = {
  description: 'update current user',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { email }: any = request.auth?.credentials;
    const allowedFields = ['slack', 'email', 'name', 'designation', 'company'];

    const user = R.pick(allowedFields, request.payload);
    return Resource.updateUserByEmail(email, user);
  }
};
