import * as Resource from './resource';

export const get = {
  description: 'get all activities',
  tags: ['api'],
  handler: async () => {
    return Resource.get();
  }
};
