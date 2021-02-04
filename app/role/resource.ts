import { getManager } from 'typeorm';
import { Role } from '../../model/role';

export const get = async (attributes: string[]) => {
  const RoleRepository = getManager().getRepository(Role);
  return RoleRepository.createQueryBuilder('role')
    .where('role.attributes @> :attributes', {
      attributes: JSON.stringify(attributes)
    })
    .getMany();
};
