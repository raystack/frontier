import Hapi from '@hapi/hapi';
import Joi from 'joi';
import * as Resource from './resource';
import { UsersResponse, UserResponse, createPayload } from './schema';
import {
  BadRequestResponse,
  InternalServerErrorResponse,
  NotFoundResponse,
  UnauthorizedResponse
} from '../../utils/schema';

export const create = {
  description: 'create user',
  tags: ['api', 'user'],
  validate: {
    payload: createPayload
  },
  response: {
    status: {
      200: UserResponse,
      400: BadRequestResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    return Resource.create(request.payload);
  }
};

export const list = {
  description: 'get list of users',
  tags: ['api', 'user'],
  validate: {
    query: Joi.object({
      action: Joi.string().optional(),
      role: Joi.string().optional()
    }).unknown(true)
  },
  response: {
    status: {
      200: UsersResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    return Resource.list(request.query);
  }
};

export const get = {
  description: 'get user by id',
  tags: ['api', 'user'],
  validate: {
    params: Joi.object({
      userId: Joi.string().required().description('user id')
    })
  },
  response: {
    status: {
      200: UserResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    return Resource.get(request.params.userId);
  }
};
