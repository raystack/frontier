import Joi from 'joi';
import Config from '../../../config/config';

export const payloadSchema = Joi.array()
  .label('GroupUserMappingCreatePayload')
  .items(
    Joi.object().keys({
      operation: Joi.string().required(),
      subject: Joi.object().required(),
      resource: Joi.object().required(),
      action: Joi.object().required()
    })
  )
  .options(Config.get('/validationOptions'));
