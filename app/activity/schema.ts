import Joi from 'joi';
import Config from '../../config/config';

const validationOptions = Config.get('/validationOptions');

export const ActivityPayload = Joi.object()
  .label('ActivityPayload')
  .keys({
    id: Joi.string().required(),
    title: Joi.string().required(),
    team: Joi.string().required(),
    details: Joi.array().items(Joi.object().optional()),
    createdAt: Joi.date().iso().required(),
    createdBy: Joi.string().required()
  })
  .options(validationOptions);
