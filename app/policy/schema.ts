import Joi from 'joi';
import Config from '../../config/config';

export const policySchema = Joi.object()
  .keys({
    subject: Joi.object({
      user: Joi.string(),
      group: Joi.string()
    })
      .xor('user', 'group')
      .required(),
    resource: Joi.object().required(),
    action: Joi.object({
      action: Joi.string(),
      role: Joi.string()
    })
      .xor('action', 'role')
      .required()
  })
  .unknown(true);

export const policiesSchema = Joi.array().items(
  policySchema.keys({
    operation: Joi.string().valid('create', 'delete').required()
  })
);

export const payloadSchema = Joi.object()
  .label('PolciesOperationPayload')
  .keys({
    policies: policiesSchema
  })
  .options(Config.get('/validationOptions'));
