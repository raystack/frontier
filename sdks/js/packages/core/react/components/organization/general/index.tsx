'use client';

import { Button, Flex, Separator, Text, Tooltip } from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { styles } from '../styles';
import { GeneralProfile } from './general.profile';
import { GeneralOrganization } from './general.workspace';
import Skeleton from 'react-loading-skeleton';
import { AuthTooltipMessage } from '~/react/utils';

export default function GeneralSetting() {
  const { activeOrganization: organization, isActiveOrganizationLoading } =
    useFrontier();

  const resource = `app/organization:${organization?.id}`;

  const listOfPermissionsToCheck = useMemo(() => {
    return [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource: resource
      },
      {
        permission: PERMISSIONS.DeletePermission,
        resource: resource
      }
    ];
  }, [resource]);

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const { canUpdateWorkspace, canDeleteWorkspace } = useMemo(() => {
    return {
      canUpdateWorkspace: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      ),
      canDeleteWorkspace: shouldShowComponent(
        permissions,
        `${PERMISSIONS.DeletePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const isLoading = isActiveOrganizationLoading || isPermissionsFetching;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>General</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <GeneralProfile organization={organization} isLoading={isLoading} />
        <Separator />
        <GeneralOrganization
          organization={organization}
          canUpdateWorkspace={canUpdateWorkspace}
          isLoading={isLoading}
        />
        <Separator />
        <GeneralDeleteOrganization
          isLoading={isLoading}
          canDelete={canDeleteWorkspace}
        />
      </Flex>
    </Flex>
  );
}

export const GeneralDeleteOrganization = ({
  isLoading,
  canDelete
}: {
  isLoading?: boolean;
  canDelete: boolean;
}) => {
  const navigate = useNavigate({ from: '/' });
  return (
    <>
      <Flex direction="column" gap="medium">
        {isLoading ? (
          <Skeleton height={'16px'} />
        ) : (
          <Text size={3} style={{ color: 'var(--foreground-muted)' }}>
            If you want to permanently delete this organization and all of its
            data.
          </Text>
        )}
        {isLoading ? (
          <Skeleton height={'32px'} width={'64px'} />
        ) : (
          <Tooltip disabled={canDelete} message={AuthTooltipMessage}>
            <Button
              variant="danger"
              type="submit"
              size="medium"
              onClick={() => navigate({ to: '/delete' })}
              disabled={!canDelete}
            >
              Delete organization
            </Button>
          </Tooltip>
        )}
        <Outlet />
      </Flex>
      <Separator />
    </>
  );
};
