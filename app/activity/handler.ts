import Hapi from '@hapi/hapi';
import * as Resource from './resource';

// groups/{groupId}/activities

// activities?{}

export const get = {
  description: 'get all activities or team activities',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { team } = request.query;
    return Resource.get(team);
  }
};
