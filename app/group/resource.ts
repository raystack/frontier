import { Group } from '../../model/group';
import * as PolicyResource from '../policy/resource';

type JSObj = Record<string, unknown>;

export const list = async (filters?: JSObj) => {
  return PolicyResource.getSubjecListWithPolicies('group', filters);
};

export const get = async (id: number, filters?: JSObj) => {
  const group = await Group.findOne(id);
  const subject = { group: group?.id };
  const policies = await PolicyResource.getPoliciesBySubject(subject, filters);
  return { ...group, policies };
};

export const create = async (payload: any, subject: JSObj) => {
  const { policies = [], ...groupPayload } = payload;
  const groupResult = await Group.save(groupPayload);
  const policyResult = await PolicyResource.bulkOperation(policies, subject);
  return { ...groupResult, policies: policyResult };
};

export const update = async (id: number, payload: any, subject: JSObj) => {
  const { policies = [], ...groupPayload } = payload;
  const group = await get(id);
  const groupResult = await Group.save({ ...group, ...groupPayload, id });
  const policyResult = await PolicyResource.bulkOperation(policies, subject);
  return { ...groupResult, policies: policyResult };
};
