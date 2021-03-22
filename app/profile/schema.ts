import Joi from 'joi';
import Config from '../../config/config';

const validationOptions = Config.get('/validationOptions');

export const updatePayload = Joi.object()
  .label('ProfileUpdatePayload')
  .keys({
    displayname: Joi.string().required(),
    metadata: Joi.object()
  })
  .unknown(true)
  .options(validationOptions);
