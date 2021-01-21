import Joi from 'joi';
import Config from '../../config/config';

const validationOptions = Config.get('/validationOptions');

export const createPayload = Joi.object()
  .label('GroupCreatePayload')
  .keys({
    name: Joi.string().required(),
    title: Joi.string().required(),
    entity: Joi.string().required(),
    privacy: Joi.string().required(),
    email: Joi.string().optional(),
    description: Joi.string().optional(),
    slack: Joi.string().optional(),
    product_group: Joi.string().optional()
  })
  .options(validationOptions);

export const updatePayload = Joi.object()
  .label('GroupUpdatePayload')
  .keys({
    name: Joi.string().required(),
    title: Joi.string().optional(),
    entity: Joi.string().optional(),
    privacy: Joi.string().optional(),
    email: Joi.string().optional(),
    description: Joi.string().optional(),
    slack: Joi.string().optional(),
    product_group: Joi.string().optional()
  })
  .options(validationOptions);
