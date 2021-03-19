import Joi from 'joi';
import Config from '../../config/config';

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
    createdAt: Joi.date().iso().required(),
    updatedAt: Joi.date().iso().required()
  })
  .label('User');

export const UsersResponse = Joi.alternatives().try(
  Joi.array().items(UserResponse).label('Users'),
  UserResponse
);

export const GroupsPolicies = Joi.array()
  .items(
    Joi.object()
      .keys({
        policies: Joi.array()
          .items(
            Joi.object()
              .keys({
                subject: Joi.object().required(),
                resource: Joi.object().required(),
                action: Joi.object().required()
              })
              .label('Policy')
          )
          .label('Policies'),
        attributes: Joi.object().optional()
      })
      .label('GroupPolicy')
  )
  .label('GroupPolicies');
