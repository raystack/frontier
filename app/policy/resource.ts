import CasbinSingleton from '../../lib/casbin';

export interface PolicyOperation {
  operation: 'delete' | 'create';
  subject: Record<string, unknown>;
  resource: Record<string, unknown>;
  action: Record<string, unknown>;
}

export const bulkOperation = async (
  policyOperations: PolicyOperation[] = [],
  subject: Record<string, unknown>
) => {
  const promiseList = policyOperations.map(async ({ operation, ...policy }) => {
    // ? the subject who is performing the action should have iam.manage permission
    const hasAccess = await CasbinSingleton.enforcer?.enforceJson(
      subject,
      policy.resource,
      { action: 'iam.manage' }
    );
    if (!hasAccess) return false;

    switch (operation) {
      case 'create': {
        await CasbinSingleton.enforcer?.addJsonPolicy(
          policy.subject,
          policy.resource,
          policy.action
        );
        break;
      }
      case 'delete': {
        await CasbinSingleton.enforcer?.removeJsonPolicy(
          policy.subject,
          policy.resource,
          policy.action
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

  return policyOperations.map((policyOperation, index) => ({
    ...policyOperation,
    success: result[index]
  }));
};
