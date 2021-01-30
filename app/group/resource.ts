import { Group } from '../../model/group';
import * as PolicyResource from '../policy/resource';

export const get = async (id: number) => {
  return Group.findOne(id);
};

export const create = async (
  payload: any,
  subject: Record<string, unknown>
) => {
  const { policies = [], ...groupPayload } = payload;
  const groupResult = await Group.save(groupPayload);
  const policyResult = await PolicyResource.bulkOperation(policies, subject);
  return { ...groupResult, policies: policyResult };
};

export const update = async (
  id: number,
  payload: any,
  subject: Record<string, unknown>
) => {
  const { policies = [], ...groupPayload } = payload;
  const group = await get(id);
  const groupResult = await Group.save({ ...group, ...groupPayload, id });
  const policyResult = await PolicyResource.bulkOperation(policies, subject);
  return { ...groupResult, policies: policyResult };
};
