import Hapi from '@hapi/hapi';
import * as Schema from './schema';
import * as Resource from './resource';

export const operation = {
  description: 'group and user mapping for create/update/delete resource',
  tags: ['api'],
  validate: {
    payload: Schema.payloadSchema
  },
  handler: async (request: Hapi.Request) => {
    const { groupIdentifier, userIdentifier } = request.params;
    return Resource.operation(groupIdentifier, userIdentifier, request.payload);
  }
};
