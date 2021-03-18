import Hapi from '@hapi/hapi';
import * as Resource from './resource';
import * as Schema from './schema';

export const get = {
  description: 'fetch roles based on attributes',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { attributes = [] } = request.query;
    return Resource.get(
      attributes.constructor === String ? [attributes] : attributes
    );
  }
};

export const create = {
  description: 'create roles along with action mapping',
  tags: ['api'],
  validate: {
    payload: Schema.createPayload
  },
  handler: async (request: Hapi.Request) => {
    const user: any = request.auth?.credentials;
    return Resource.create(<any>request.payload, user);
  }
};
