import Hapi from '@hapi/hapi';
import * as Resource from './resource';

export const selfList = {
  description: 'fetch list of groups of the loggedIn user',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { id: loggedInUserId } = request?.auth?.credentials || { id: '' };
    return Resource.list(<string>loggedInUserId, request.query);
  }
};

export const list = {
  description: 'fetch list of groups of a user',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { userId } = request.params;
    return Resource.list(userId, request.query);
  }
};
