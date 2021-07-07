import { getManager } from 'typeorm';
import * as R from 'ramda';
import Boom from '@hapi/boom';
import { Role } from '../../model/role';
import CasbinSingleton from '../../lib/casbin';

type JSObj = Record<string, unknown>;
interface ActionOperation {
  operation: 'create' | 'delete';
  action: string;
}
interface RolePayload {
  displayname: string;
  attributes?: string[];
  metadata?: Record<string, unknown>;
  actions?: ActionOperation[];
}

interface GetQuery {
  tags?: string[];
}
export const get = async (attributes: string[], query: GetQuery = {}) => {
  const RoleRepository = getManager().getRepository(Role);
  const { tags = [] } = query;
  const cursor = RoleRepository.createQueryBuilder('role').where(
    'role.attributes @> :attributes',
    {
      attributes: JSON.stringify(attributes)
    }
  );

  if (tags.length > 0) {
    cursor.andWhere('role.tags @> :tags', {
      tags
    });
  }
  return cursor.getMany();
};

export const mapActionRoleInBulk = async (
  roleId: string,
  actionOperations: ActionOperation[] = [],
  loggedInUser: any
) => {
  const promiseList = actionOperations.map(async ({ operation, action }) => {
    const options: JSObj = { created_by: loggedInUser };
    switch (operation) {
      case 'create': {
        // TODO: Check why .addActionGroupingJsonPolicy is not callable anymore
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        await CasbinSingleton.enforcer?.addActionGroupingJsonPolicy(
          { action },
          { role: roleId },
          options
        );
        break;
      }
      case 'delete': {
        // TODO: Check why .removeActionGroupingJsonPolicy is not callable anymore
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        await CasbinSingleton.enforcer?.removeActionGroupingJsonPolicy(
          { action },
          { role: roleId },
          options
        );
        break;
      }
      default: {
        break;
      }
    }
    return true;
  });

  const result = await Promise.all(promiseList);

  return actionOperations.map((actionOperation, index) => ({
    ...actionOperation,
    success: result[index]
  }));
};

export const create = async (payload: RolePayload, loggedInUser: any) => {
  const { actions = [], ...rolePayload } = payload;
  const role = await Role.save(<any>rolePayload, {
    data: { user: loggedInUser }
  });

  if (!R.isEmpty(actions)) {
    await mapActionRoleInBulk(role.id, actions, loggedInUser);
  }

  return role;
};

export const update = async (
  roleId: string,
  payload: RolePayload,
  loggedInUser: any
) => {
  const { actions = [], ...rolePayload } = payload;
  const role = await Role.findOne(roleId);
  if (!role) return Boom.notFound('Role not found');

  const rolePayloadToUpdate: any = {
    ...role,
    ...rolePayload
  };
  const updatedRole = await Role.save(rolePayloadToUpdate, {
    data: { user: loggedInUser }
  });

  if (!R.isEmpty(actions)) {
    await mapActionRoleInBulk(role.id, actions, loggedInUser);
  }

  return updatedRole;
};
