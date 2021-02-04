import Hapi from '@hapi/hapi';
import * as Resource from './resource';

export const get = {
  description: 'fetch roles based on attributes',
  tags: ['api'],
  handler: async (request: Hapi.Request) => {
    const { attributes } = request.params;
    return Resource.get(
      attributes.constructor === String ? [attributes] : attributes
    );
  }
};
