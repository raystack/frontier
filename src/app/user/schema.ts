import Joi from 'joi';
import Config from '../../config/config';
import { policiesSchema } from '../policy/schema';

const validationOptions = Config.get('/validationOptions');

export const createPayload = Joi.object()
  .label('UserCreatePayload')
  .keys({
    displayname: Joi.string().required(),
    metadata: Joi.object()
  })
  .options(validationOptions);

export const UserResponse = Joi.object()
  .keys({
    id: Joi.string().uuid().required(),
    username: Joi.string().required(),
    displayname: Joi.string().required(),
    metadata: Joi.object(),
    createdAt: Joi.date().iso().optional(),
    updatedAt: Joi.date().iso().optional()
  })
  .unknown(true)
  .label('User');

export const UsersResponse = Joi.alternatives().try(
  Joi.array().items(UserResponse).label('Users'),
  UserResponse
);

export const GroupsPolicies = Joi.array()
  .items(
    Joi.object()
      .keys({
        policies: policiesSchema,
        attributes: Joi.array().items(Joi.object())
      })
      .unknown(true)
      .label('GroupPolicy')
  )
  .label('GroupPolicies');
