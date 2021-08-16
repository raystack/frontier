import Hapi from '@hapi/hapi';
import Joi from 'joi';
import * as Resource from './resource';
import * as PolicySchema from '../../policy/schema';
import {
  InternalServerErrorResponse,
  NotFoundResponse,
  UnauthorizedResponse
} from '../../../utils/schema';
import {
  UserGroupMapping,
  UserWithPoliciesResponse,
  UsersWithPoliciesResponse
} from './schema';
import { clearQueryCache } from '../../../utils/cache';

const iamConfig = (actionName: string) => ({
  iam: {
    permissions: [
      {
        action: actionName,
        attributes: [{ group: { key: 'groupId', type: 'params' } }]
      }
    ]
  }
});

export const list = {
  description: 'fetch list of users of a group',
  tags: ['api', 'group'],
  validate: {
    query: Joi.object({
      fields: Joi.object()
    }).unknown(true),
    params: Joi.object({
      groupId: Joi.string().required().description('group id')
    })
  },
  response: {
    status: {
      200: UsersWithPoliciesResponse,
      401: UnauthorizedResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { groupId } = request.params;
    return Resource.list(groupId, request.query);
  }
};

export const get = {
  description: 'fetch user and group mapping',
  tags: ['api', 'group'],
  validate: {
    query: Joi.object({
      role: Joi.string().optional(),
      action: Joi.string().optional(),
      fields: Joi.object().optional()
    }).unknown(true),
    params: Joi.object({
      groupId: Joi.string().required().description('group id'),
      userId: Joi.string().required().description('user id')
    })
  },
  response: {
    status: {
      200: UserWithPoliciesResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { groupId, userId } = request.params;
    return Resource.get(groupId, userId, request.query);
  }
};

export const post = {
  description: 'create group and user mapping',
  tags: ['api', 'group'],
  validate: {
    payload: PolicySchema.payloadSchema,
    params: Joi.object({
      groupId: Joi.string().required().description('group id'),
      userId: Joi.string().required().description('user id')
    })
  },
  response: {
    status: {
      200: UserGroupMapping,
      201: UserGroupMapping,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  app: iamConfig('iam.create'),
  handler: async (request: Hapi.Request) => {
    const { payload } = request;
    const { groupId, userId } = request.params;
    const { id: loggedInUserId } = request.auth.credentials;
    await Promise.all([
      clearQueryCache(request, <string>loggedInUserId),
      clearQueryCache(request, <string>userId)
    ]);
    return Resource.create(groupId, userId, loggedInUserId, payload);
  }
};

export const put = {
  description: 'update group and user mapping',
  tags: ['api', 'group'],
  validate: {
    payload: PolicySchema.payloadSchema,
    params: Joi.object({
      groupId: Joi.string().required().description('group id'),
      userId: Joi.string().required().description('user id')
    })
  },
  response: {
    status: {
      200: UserGroupMapping,
      201: UserGroupMapping,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  app: iamConfig('iam.manage'),
  handler: async (request: Hapi.Request) => {
    const { payload } = request;
    const { groupId, userId } = request.params;
    const { id: loggedInUserId } = request.auth.credentials;
    await Promise.all([
      clearQueryCache(request, <string>loggedInUserId),
      clearQueryCache(request, <string>userId)
    ]);

    return Resource.update(groupId, userId, loggedInUserId, payload);
  }
};

export const remove = {
  description: 'delete group and user mapping',
  tags: ['api', 'group'],
  validate: {
    params: Joi.object({
      groupId: Joi.string().required().description('group id'),
      userId: Joi.string().required().description('user id')
    })
  },
  response: {
    status: {
      200: Joi.bool(),
      201: UserGroupMapping,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  app: iamConfig('iam.delete'),
  handler: async (request: Hapi.Request) => {
    const { groupId, userId } = request.params;
    const { id: loggedInUserId } = request.auth.credentials;
    await Promise.all([
      clearQueryCache(request, <string>loggedInUserId),
      clearQueryCache(request, <string>userId)
    ]);
    return Resource.remove(groupId, userId, <string>loggedInUserId);
  }
};

export const removeSelf = {
  description: 'delete group and loggedin user mapping',
  tags: ['api', 'group'],
  validate: {
    params: Joi.object({
      groupId: Joi.string().required().description('group id')
    })
  },
  response: {
    status: {
      200: Joi.bool(),
      201: UserGroupMapping,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { groupId } = request.params;
    const { id: loggedInUserId } = request.auth.credentials;
    await clearQueryCache(request, <string>loggedInUserId);
    return Resource.remove(
      groupId,
      <string>loggedInUserId,
      <string>loggedInUserId
    );
  }
};
