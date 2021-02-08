import Joi from 'joi';
import Config from '../../../config/config';

export const policiesSchema = Joi.array().items(
  Joi.object().keys({
    operation: Joi.string().required(),
    subject: Joi.object().required(),
    resource: Joi.object().required(),
    action: Joi.object().required()
  })
);

export const payloadSchema = Joi.object()
  .label('GroupUserMappingCreatePayload')
  .keys({
    policies: policiesSchema
  })
  .options(Config.get('/validationOptions'));
