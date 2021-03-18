import { getManager } from 'typeorm';
import { Role } from '../../model/role';
import CasbinSingleton from '../../lib/casbin';

interface ROLE_PAYLOAD {
  displayname: string;
  attributes?: string[];
  metadata?: Record<string, unknown>;
  actions?: string[];
}

export const get = async (attributes: string[]) => {
  const RoleRepository = getManager().getRepository(Role);
  return RoleRepository.createQueryBuilder('role')
    .where('role.attributes @> :attributes', {
      attributes: JSON.stringify(attributes)
    })
    .getMany();
};

export const create = async (payload: ROLE_PAYLOAD, loggedInUser: any) => {
  const { actions = [], ...rolePayload } = payload;
  const role = await Role.save(<any>rolePayload, {
    data: { user: loggedInUser }
  });

  const actionRolePromiseList = actions.map(async (action: string) => {
    // TODO: Check why .addActionGroupingJsonPolicy is not callable anymore
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    return CasbinSingleton.enforcer?.addActionGroupingJsonPolicy(
      { action },
      { role: role.id },
      { created_by: loggedInUser }
    );
  });
  await Promise.all(actionRolePromiseList);

  return role;
};
