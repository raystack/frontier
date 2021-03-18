import Hapi from '@hapi/hapi';
import Joi from 'joi';
import * as Resource from './resource';
import { RolesResponse, Attributes } from './schema';
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
      attributes: Attributes
    })
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
    const { attributes = [] } = request.query;
    return Resource.get(
      attributes.constructor === String ? [attributes] : attributes
    );
  }
};
