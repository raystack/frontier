import * as R from 'ramda';
import { Group } from '../../model/group';
import * as PolicyResource from '../policy/resource';

type JSObj = Record<string, unknown>;

const extractResourceAction = (data: JSObj = {}) => {
  const ACTION_KEYS = ['action', 'role'];
  const action = R.pick(ACTION_KEYS, data);
  const resource = R.omit(ACTION_KEYS, data);
  return { resource, action };
};

export const get = async (id: number, filters?: JSObj) => {
  const group = await Group.findOne(id);
  const subject = { group: group?.name };
  const { resource, action } = extractResourceAction(filters);
  const policies = await PolicyResource.list(subject, resource, action);
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
