import Hapi from '@hapi/hapi';
import * as Schema from './schema';
import * as Resource from './resource';
import { UserResponse } from '../user/schema';
import {
  InternalServerErrorResponse,
  NotFoundResponse,
  UnauthorizedResponse
} from '../../utils/schema';
import { clearQueryCache } from '../../utils/cache';

export const get = {
  description: "get current user's profile",
  tags: ['api', 'profile'],
  response: {
    status: {
      200: UserResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { id }: any = request.auth?.credentials;
    return Resource.getUserById(id);
  }
};

export const update = {
  description: 'update current user',
  tags: ['api', 'profile'],
  validate: {
    payload: Schema.updatePayload
  },
  response: {
    status: {
      200: UserResponse,
      201: UserResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { id }: any = request.auth?.credentials;
    await clearQueryCache(request, id);
    return Resource.updateUserById(id, request.payload);
  }
};
