import { useMemo } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, BatchCheckPermissionBody } from '@raystack/proton/frontier';
import { formatPermissions } from '~/utils';

export const usePermissions = (
  permissions: BatchCheckPermissionBody[] = [],
  shouldCalled: boolean | undefined = true
) => {
  const { data, isLoading } = useQuery(
    FrontierServiceQueries.batchCheckPermission,
    { bodies: permissions },
    { enabled: shouldCalled && permissions.length > 0 }
  );

  const permissionsMap = useMemo(() => {
    const pairs = data?.pairs ?? [];
    if (!pairs.length) return {};
    const normalizedPairs = pairs.map((p: any) => ({
      body: p?.body ?? {},
      status: Boolean(p?.status)
    }));
    return formatPermissions(normalizedPairs);
  }, [data?.pairs]);

  return { isFetching: isLoading, permissions: permissionsMap };
};
