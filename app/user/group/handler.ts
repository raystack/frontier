import Hapi from '@hapi/hapi';
import Joi from 'joi';
import * as Resource from './resource';
import { GroupsPolicies } from '../schema';
import {
  InternalServerErrorResponse,
  NotFoundResponse,
  UnauthorizedResponse
} from '../../../utils/schema';

export const selfList = {
  description: 'fetch list of groups of the loggedIn user',
  tags: ['api', 'user'],
  validate: {
    query: Joi.object({
      action: Joi.string().optional()
    }).unknown(true)
  },
  response: {
    status: {
      200: GroupsPolicies,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { id: loggedInUserId } = request?.auth?.credentials || { id: '' };
    return Resource.list(<string>loggedInUserId, request.query);
  }
};

export const list = {
  description: 'fetch list of groups of a user',
  tags: ['api', 'user'],
  validate: {
    params: Joi.object({
      userId: Joi.string().required().description('user id')
    }),
    query: Joi.object({
      action: Joi.string().optional()
    }).unknown(true)
  },
  response: {
    status: {
      200: GroupsPolicies,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { userId } = request.params;
    return Resource.list(userId, request.query);
  }
};
