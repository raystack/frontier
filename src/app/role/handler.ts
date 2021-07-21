import Hapi from '@hapi/hapi';
import Joi from 'joi';
import * as Resource from './resource';
import {
  RolesResponse,
  Attributes,
  createPayload,
  RoleResponse
} from './schema';
import {
  InternalServerErrorResponse,
  NotFoundResponse,
  UnauthorizedResponse
} from '../../utils/schema';

export const get = {
  description: 'get roles based on attributes',
  tags: ['api', 'role'],
  validate: {
    query: Joi.object({
      attributes: Attributes,
      tags: Joi.array().single().items(Joi.string())
    })
      .rename('tags[]', 'tags', { ignoreUndefined: true })
      .unknown(true)
  },
  response: {
    status: {
      200: RolesResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const { attributes = [], ...query } = request.query;
    return Resource.get(
      attributes.constructor === String ? [attributes] : attributes,
      query
    );
  }
};

export const create = {
  description: 'create roles along with action mapping',
  tags: ['api', 'role'],

  app: {
    iam: {
      permissions: [
        {
          action: 'role.create',
          attributes: [{ '*': { defaultValue: '*' } }]
        }
      ]
    }
  },
  validate: {
    payload: createPayload
  },
  response: {
    status: {
      200: RoleResponse,
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

export const update = {
  description: 'update roles along with action mapping',
  tags: ['api', 'role'],
  app: {
    iam: {
      permissions: [
        {
          action: 'role.manage',
          attributes: [{ '*': { defaultValue: '*' } }]
        }
      ]
    }
  },
  validate: {
    params: Joi.object().keys({
      id: Joi.string().required()
    }),
    payload: createPayload
  },
  response: {
    status: {
      200: RoleResponse,
      401: UnauthorizedResponse,
      404: NotFoundResponse,
      500: InternalServerErrorResponse
    }
  },
  handler: async (request: Hapi.Request) => {
    const user: any = request.auth?.credentials;
    return Resource.update(request.params.id, <any>request.payload, user);
  }
};
