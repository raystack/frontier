import Joi from 'joi';
import Config from '../../config/config';

const validationOptions = Config.get('/validationOptions');

const Role = Joi.object()
  .keys({
    id: Joi.string().uuid().required(),
    displayname: Joi.string().required(),
    attributes: Joi.array().items(Joi.string()),
    metadata: Joi.object().required(),
    createdAt: Joi.date().iso().required(),
    updatedAt: Joi.date().iso().required()
  })
  .label('Role')
  .options(validationOptions);

export const RolesResponse = Joi.array()
  .items(Role)
  .label('Roles')
  .options(validationOptions);

export const Attributes = Joi.alternatives().try(
  Joi.array().items(Joi.string().optional()).label('attributes').optional(),
  Joi.string().optional()
);
