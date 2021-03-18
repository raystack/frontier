import Joi from 'joi';
import Config from '../../config/config';

const validationOptions = Config.get('/validationOptions');

export const createPayload = Joi.object()
  .label('RoleCreatePayload')
  .keys({
    displayname: Joi.string().required(),
    attributes: Joi.array().items(Joi.string()),
    actions: Joi.array().items(Joi.string()),
    metadata: Joi.object()
  })
  .options(validationOptions);
