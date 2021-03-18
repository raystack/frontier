import Hapi from '@hapi/hapi';
import * as Resource from './resource';
import { CheckAccessResponse, checkAccessPayload } from './schema';
import {
  InternalServerErrorResponse,
  NotFoundResponse,
  UnauthorizedResponse
} from '../../utils/schema';

export const checkAccess = {
  description:
    'checks whether subject has access to perform action on a given resource/attributes',
  tags: ['api', 'check-access'],
  validate: {
    payload: checkAccessPayload
  },
  response: {
    status: {
      200: CheckAccessResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    return Resource.checkAccess(<any>request.payload);
  }
};
