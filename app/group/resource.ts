import { Group } from '../../model/group';

export const get = async (id: number) => {
  return Group.findOne(id);
};

export const create = async (payload: any) => {
  return Group.save(payload);
};

export const update = async (id: number, payload: any) => {
  await Group.update({ id }, payload);
  return get(id);
};
