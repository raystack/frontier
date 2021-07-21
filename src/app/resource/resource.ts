import CasbinSingleton from '../../lib/casbin';

type JSObj = Record<string, unknown>;

interface ResourceAttributeMappingOperation {
  operation: 'create' | 'delete';
  resource: JSObj;
  attributes: JSObj;
}

type ResourceAttributeMappingOperationList = ResourceAttributeMappingOperation[];

const handleCreateOperation = async (
  resource: JSObj,
  attributes: JSObj,
  loggedInUser: any
) => {
  const hasAttributeAccess = await CasbinSingleton.enforcer?.enforceJson(
    { user: loggedInUser.id },
    attributes,
    { action: 'resource.create' }
  );
  if (!hasAttributeAccess) return false;

  if (hasAttributeAccess) {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    await CasbinSingleton.enforcer?.addResourceGroupingJsonPolicy(
      resource,
      attributes,
      { created_by: loggedInUser }
    );
    return true;
  }

  return false;
};

const handleDeleteOperation = async (
  resource: JSObj,
  attributes: JSObj,
  loggedInUser: any
) => {
  const hasAttributeAccess = await CasbinSingleton.enforcer?.enforceJson(
    { user: loggedInUser.id },
    attributes,
    { action: 'resource.delete' }
  );
  const hasResourceAccess = await CasbinSingleton.enforcer?.enforceJson(
    { user: loggedInUser.id },
    resource,
    { action: 'resource.delete' }
  );
  if (hasAttributeAccess && hasResourceAccess) {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    await CasbinSingleton.enforcer?.removeResourceGroupingJsonPolicy(
      resource,
      attributes,
      { created_by: loggedInUser }
    );

    return true;
  }

  return false;
};

export const create = async (
  resourceAttributeMappingOperationList: ResourceAttributeMappingOperationList,
  loggedInUser: any
) => {
  const promiseList = resourceAttributeMappingOperationList.map(
    async (operationObj: ResourceAttributeMappingOperation) => {
      const { operation, resource, attributes } = operationObj;
      if (operation === 'create') {
        return handleCreateOperation(resource, attributes, loggedInUser);
      }
      if (operation === 'delete') {
        return handleDeleteOperation(resource, attributes, loggedInUser);
      }

      return false;
    }
  );

  const result = await Promise.all(promiseList);

  return resourceAttributeMappingOperationList.map((operationObj, index) => ({
    ...operationObj,
    success: result[index]
  }));
};
