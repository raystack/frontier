import * as Resource from './resource';
import { ActivityPayloadSuccessResponse } from './schema';
import {
  InternalServerErrorResponse,
  NotFoundResponse,
  UnauthorizedResponse
} from '../../utils/schema';

export const get = {
  description: 'get all activities',
  tags: ['api', 'activity'],
  response: {
    status: {
      200: ActivityPayloadSuccessResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async () => {
    return await Resource.get();
  }
};
