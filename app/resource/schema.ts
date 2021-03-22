import Joi from 'joi';
import Config from '../../config/config';

const validationOptions = Config.get('/validationOptions');

const singleOperationSchema = Joi.object()
  .keys({
    operation: Joi.string().valid('create', 'delete').required(),
    resource: Joi.object(),
    attributes: Joi.object()
  })
  .unknown(true);

export const createPayload = Joi.array()
  .label('ResourceAttributesMappingPayload')
  .items(singleOperationSchema)
  .options(validationOptions);

export const createPayloadResponse = Joi.array()
  .label('ResourceAttributesMappingPayloadResponse')
  .items(singleOperationSchema.keys({ success: Joi.boolean() }))
  .options(validationOptions);
