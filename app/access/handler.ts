import Hapi from '@hapi/hapi';
import * as Resource from './resource';
import * as Schema from './schema';

export const checkAccess = {
  description:
    'checks whether subject has access to perform action on a given resource/attributes',
  tags: ['api'],
  validate: {
    payload: Schema.checkAccessPayload
  },
  handler: async (request: Hapi.Request) => {
    return Resource.checkAccess(<any>request.payload);
  }
};
