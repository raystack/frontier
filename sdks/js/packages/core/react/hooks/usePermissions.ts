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

  const { client, activeOrganization: organization } = useFrontier();

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [client]);

  useEffect(() => {
    if (shouldCalled) {
      fetchOrganizationPermissions();
    }
  }, [fetchOrganizationPermissions, shouldCalled]);

  const permissionsMap = useMemo(() => {
    if (permisionValues.length) {
      return formatPermissions(permisionValues);
    } else {
      return {};
    }
  }, [permisionValues]);

  return { isFetching: fetchingPermissions, permissions: permissionsMap };
};
