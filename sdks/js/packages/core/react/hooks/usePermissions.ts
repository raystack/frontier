import { useMemo } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries,
  BatchCheckPermissionBodySchema,
  BatchCheckPermissionResponsePair,
  BatchCheckPermissionRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { formatPermissions } from '~/utils';

interface PermissionCheck {
  permission: string;
  resource: string;
}

export const usePermissions = (
  permissions: PermissionCheck[] = [],
  shouldCalled: boolean | undefined = true
): { isFetching: boolean; permissions: Record<string, boolean>; error?: unknown } => {
  const protobufPermissions = useMemo(() => {
    return permissions.map(permission => 
      create(BatchCheckPermissionBodySchema, permission)
    );
  }, [permissions]);

  const { data, isLoading, error } = useQuery(
    FrontierServiceQueries.batchCheckPermission,
    create(BatchCheckPermissionRequestSchema, { bodies: protobufPermissions }),
    { enabled: shouldCalled && permissions.length > 0 }
  );

  const permissionsMap = useMemo(() => {
    const pairs = data?.pairs ?? [];
    if (!pairs.length) return {};
    const normalizedPairs = pairs.map((p: BatchCheckPermissionResponsePair) => ({
      body: p?.body ?? {},
      status: Boolean(p?.status)
    }));
    return formatPermissions(normalizedPairs);
  }, [data?.pairs]);

  return { isFetching: isLoading, permissions: permissionsMap, error };
};
