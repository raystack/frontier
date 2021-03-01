import Joi from 'joi';
import Config from '../../config/config';
import * as PolicySchema from '../policy/schema';

const validationOptions = Config.get('/validationOptions');

export const createPayload = Joi.object()
  .label('GroupCreatePayload')
  .keys({
    groupname: Joi.string().optional(),
    displayname: Joi.string().required(),
    policies: PolicySchema.policiesSchema,
    attributes: Joi.array().items(Joi.object()),
    metadata: Joi.object()
  })
  .options(validationOptions);

export const updatePayload = Joi.object()
  .label('GroupUpdatePayload')
  .keys({
    displayname: Joi.string().required(),
    policies: PolicySchema.policiesSchema,
    attributes: Joi.array().items(Joi.object()),
    metadata: Joi.object()
  })
  .options(validationOptions);
