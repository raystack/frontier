import Hapi from '@hapi/hapi';
import * as Resource from './resource';
import {
  InternalServerErrorResponse,
  NotFoundResponse,
  UnauthorizedResponse
} from '../../utils/schema';
import { createPayload, createPayloadResponse } from './schema';

export const create = {
  description: 'create resource and attributes mapping',
  tags: ['api', 'resource'],
  validate: {
    payload: createPayload
  },
  response: {
    status: {
      200: createPayloadResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const user: any = request.auth?.credentials;
    return Resource.create(<any>request.payload, user);
  }
};
