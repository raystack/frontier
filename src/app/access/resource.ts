import CasbinSingleton from '../../lib/casbin';

type AccessObj = {
  subject: {
    user?: string;
    group?: string;
  };
  resource: Record<string, unknown>;
  action: {
    action?: string;
    role?: string;
  };
};
export type AccessList = AccessObj[];

export const checkAccess = async (accessList: AccessList = []) => {
  const result = await CasbinSingleton.enforcer?.batchEnforceJson(accessList);

  return accessList.map((accessObj, index) => ({
    ...accessObj,
    hasAccess: result && result[index]
  }));
};
