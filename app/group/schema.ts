import Joi from 'joi';
import Config from '../../config/config';
import * as PolicySchema from '../policy/schema';

const validationOptions = Config.get('/validationOptions');

export const createPayload = Joi.object()
  .label('GroupCreatePayload')
  .keys({
    groupname: Joi.string().optional(),
    displayname: Joi.string().required(),
    policies: PolicySchema.policiesSchema,
    attributes: Joi.array().items(Joi.object()),
    metadata: Joi.object()
  })
  .options(validationOptions);

export const updatePayload = Joi.object()
  .label('GroupUpdatePayload')
  .keys({
    displayname: Joi.string().required(),
    policies: PolicySchema.policiesSchema,
    attributes: Joi.array().items(Joi.object()),
    metadata: Joi.object()
  })
  .options(validationOptions);

export const GroupPolicies = Joi.object()
  .keys({
    isMember: Joi.bool().required(),
    userPolicies: Joi.array()
      .items(
        Joi.object()
          .keys({
            subject: Joi.string().required(),
            resource: Joi.string().required(),
            action: Joi.string().required()
          })
          .label('Policy')
      )
      .label('Policies'),
    memberCount: Joi.number().integer().required(),
    attributes: Joi.array().items(Joi.object())
  })
  .label('GroupPolicy');

export const GroupsPolicies = Joi.array()
  .items(GroupPolicies)
  .label('GroupsPolicies');
