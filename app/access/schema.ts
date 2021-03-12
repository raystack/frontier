import Joi from 'joi';
import Config from '../../config/config';
import * as PolicySchema from '../policy/schema';

const validationOptions = Config.get('/validationOptions');

export const checkAccessPayload = Joi.array()
  .label('CheckAccessPayload')
  .items(PolicySchema.policySchema)
  .options(validationOptions);
