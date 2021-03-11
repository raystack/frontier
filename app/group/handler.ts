import Hapi from '@hapi/hapi';
import * as Schema from './schema';
import * as Resource from './resource';
import { get as ActivitiesByGroup } from '../activity/resource';

export const list = {
  description: 'get list of groups',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { id: loggedInUserId } = request?.auth?.credentials || { id: '' };
    return Resource.list(request.query, <string>loggedInUserId);
  }
};

export const get = {
  description: 'get group by id',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { id: loggedInUserId } = request?.auth?.credentials || { id: '' };
    return Resource.get(
      request.params.id,
      <string>loggedInUserId,
      request.query
    );
  }
};

export const create = {
  description: 'create group',
  tags: ['api'],
  validate: {
    payload: Schema.createPayload
  },
  handler: async (request: Hapi.Request) => {
    const { id: loggedInUserId } = request?.auth?.credentials || { id: '' };
    return Resource.create(request.payload, <string>loggedInUserId);
  }
};

export const update = {
  description: 'update group by id',
  tags: ['api'],
  validate: {
    payload: Schema.updatePayload
  },
  handler: async (request: Hapi.Request) => {
    const { id: loggedInUserId } = request?.auth?.credentials || { id: '' };
    return Resource.update(
      request.params.id,
      request.payload,
      <string>loggedInUserId
    );
  }
};

export const getActivitiesByGroup = {
  description: 'get activities by group',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { id: groupId } = request.params;
    return ActivitiesByGroup(groupId);
  }
};
