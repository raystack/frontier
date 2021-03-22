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
  .unknown(true)
  .label('Role')
  .options(validationOptions);

export const RoleResponse = Role;
export const RolesResponse = Joi.array()
  .items(Role)
  .label('Roles')
  .options(validationOptions);

export const Attributes = Joi.alternatives().try(
  Joi.array().items(Joi.string().optional()).label('attributes').optional(),
  Joi.string().optional()
);

const actionOperationPayload = Joi.object()
  .label('ActionOperation')
  .keys({
    operation: Joi.string().valid('create', 'delete').required(),
    action: Joi.string().required()
  })
  .unknown(true);

export const createPayload = Joi.object()
  .label('RoleCreatePayload')
  .keys({
    displayname: Joi.string().required(),
    attributes: Joi.array().items(Joi.string()),
    actions: Joi.array().items(actionOperationPayload),
    metadata: Joi.object()
  })
  .options(validationOptions);
