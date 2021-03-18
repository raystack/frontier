import Hapi from '@hapi/hapi';
import Joi from 'joi';
import * as Resource from './resource';
import { get as ActivitiesByGroup } from '../activity/resource';
import {
  GroupPolicies,
  GroupsPolicies,
  createPayload,
  updatePayload
} from './schema';
import { ActivityPayloadSuccessResponse } from '../activity/schema';
import {
  BadRequestResponse,
  InternalServerErrorResponse,
  NotFoundResponse,
  UnauthorizedResponse
} from '../../utils/schema';

export const list = {
  description: 'get list of groups',
  tags: ['api', 'group'],
  validate: {
    query: Joi.object({
      user_role: Joi.string().optional().description('role id'),
      group: Joi.string().optional().description('group id')
    })
  },
  response: {
    status: {
      200: GroupsPolicies,
      401: UnauthorizedResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { id: loggedInUserId } = request?.auth?.credentials || { id: '' };
    return Resource.list(request.query, <string>loggedInUserId);
  }
};

export const get = {
  description: 'get group by id',
  tags: ['api', 'group'],
  validate: {
    query: Joi.object().keys({
      user_role: Joi.string().optional(),
      group: Joi.string().optional()
    }),
    params: Joi.object().keys({
      id: Joi.string().required().description('group id')
    })
  },
  response: {
    status: {
      200: GroupPolicies,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { id: loggedInUserId } = request?.auth?.credentials || { id: '' };
    return Resource.get(
      request.params.id,
      <string>loggedInUserId,
      request.query
    );
  }
};

export const create = {
  description: 'create group',
  tags: ['api', 'group'],
  validate: {
    payload: createPayload
  },
  response: {
    status: {
      200: createPayload,
      400: BadRequestResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { id: loggedInUserId } = request?.auth?.credentials || { id: '' };
    return Resource.create(request.payload, <string>loggedInUserId);
  }
};

export const update = {
  description: 'update group by id',
  tags: ['api', 'group'],
  validate: {
    payload: updatePayload,
    params: Joi.object().keys({
      id: Joi.string().required().description('group id')
    })
  },
  response: {
    status: {
      200: updatePayload,
      201: updatePayload,
      400: BadRequestResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { id: loggedInUserId } = request?.auth?.credentials || { id: '' };
    return Resource.update(
      request.params.id,
      request.payload,
      <string>loggedInUserId
    );
  }
};

export const getActivitiesByGroup = {
  description: 'get activities by group',
  tags: ['api', 'group'],
  validate: {
    params: Joi.object().keys({
      id: Joi.string().required().description('group id')
    })
  },
  response: {
    status: {
      200: ActivityPayloadSuccessResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { id: groupId } = request.params;
    return ActivitiesByGroup(groupId);
  }
};
