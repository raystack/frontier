import Joi from 'joi';
import Config from '../../config/config';

const validationOptions = Config.get('/validationOptions');

export const ActivityPayload = Joi.object()
  .label('Activity')
  .keys({
    id: Joi.string().required(),
    title: Joi.string().required(),
    model: Joi.string().required(),
    documentId: Joi.string().required(),
    document: Joi.object().required(),
    diffs: Joi.array().items(Joi.object().optional()),
    createdBy: Joi.string().required(),
    createdAt: Joi.date().iso().required()
  })
  .options(validationOptions);

const ActivityResponse = Joi.object()
  .keys({
    id: Joi.string().uuid(),
    createdAt: Joi.date().iso(),
    diff: Joi.object(),
    reason: Joi.string(),
    user: Joi.string()
  })
  .label('Activity');

export const ActivityPayloadSuccessResponse = Joi.array()
  .label('Activities')
  .items(ActivityResponse);
