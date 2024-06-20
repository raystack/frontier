import { useCallback, useEffect, useMemo, useState } from 'react';
import { V1Beta1BatchCheckPermissionBody } from '~/src';
import { formatPermissions } from '~/utils';
import { useFrontier } from '../contexts/FrontierContext';

export const usePermissions = (
  permissions: V1Beta1BatchCheckPermissionBody[] = [],
  shouldCalled: boolean | undefined = true
) => {
  const [permisionValues, setPermisionValues] = useState([]);
  const [fetchingPermissions, setFetchingOrgPermissions] = useState(false);

  const { client } = useFrontier();

  const fetchOrganizationPermissions = useCallback(async () => {
    try {
      setFetchingOrgPermissions(true);
      const {
        // @ts-ignore
        data: { pairs }
      } = await client?.frontierServiceBatchCheckPermission({
        bodies: permissions
      });
      setPermisionValues(pairs);
    } catch (err) {
      console.error(err);
    } finally {
      setFetchingOrgPermissions(false);
    }
  }, [client, permissions]);

  useEffect(() => {
    if (shouldCalled && permissions.length > 0) {
      fetchOrganizationPermissions();
    }
  }, [fetchOrganizationPermissions, permissions.length, shouldCalled]);

  const permissionsMap = useMemo(() => {
    if (permisionValues.length) {
      return formatPermissions(permisionValues);
    } else {
      return {};
    }
  }, [permisionValues]);

  return { isFetching: fetchingPermissions, permissions: permissionsMap };
};
