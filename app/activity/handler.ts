import Hapi from '@hapi/hapi';
import * as Resource from './resource';

// groups/{groupId}/activities

// activities?{}

export const get = {
  description: 'get all activities or group activities',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { group } = request.query;
    return Resource.get(group);
  }
};
