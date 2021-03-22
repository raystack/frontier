import Joi from 'joi';
import { UserResponse } from '../../user/schema';
import { policySchema } from '../../policy/schema';

const UserWithPolicies = UserResponse.append({
  policies: Joi.array().items(
    Joi.object()
      .keys({
        subject: Joi.object(),
        resource: Joi.object(),
        action: Joi.object()
      })
      .unknown(true)
      .label('Policy')
  )
}).label('UserWithPolicies');

export const UserWithPoliciesResponse = UserWithPolicies;

export const UserGroupMapping = Joi.array()
  .items(
    policySchema.keys({
      operation: Joi.string().valid('create', 'delete').required(),
      success: Joi.bool().required()
    })
  )
  .label('UserGroupResponse');
