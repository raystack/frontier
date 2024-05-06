import { useMemo } from 'react';
import { useFrontier } from '../contexts/FrontierContext';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { usePermissions } from './usePermissions';

export const useBillingPermission = () => {
  const { activeOrganization } = useFrontier();

  const resource = `app/organization:${activeOrganization?.id}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.DeletePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!activeOrganization?.id
  );

  const { isAllowed } = useMemo(() => {
    return {
      isAllowed: shouldShowComponent(
        permissions,
        `${PERMISSIONS.DeletePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  return {
    isFetching,
    isAllowed
  };
};
