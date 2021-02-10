import { User } from '../../model/user';

export const create = async (payload: any) => {
  return await User.save(payload);
};
