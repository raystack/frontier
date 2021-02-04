import Hapi from '@hapi/hapi';
import * as Resource from './resource';

export const get = {
  description: 'fetch roles for user',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    return Resource.get();
  }
};
