'use client';

import { useMemo } from 'react';
import {
  Button,
  Tooltip,
  Separator,
  Skeleton,
  Text,
  Flex
} from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { usePermissions } from '~/react/hooks/usePermissions';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { GeneralOrganization } from './general.workspace';
import { AuthTooltipMessage } from '~/react/utils';
import { styles } from '../styles';
import { useTerminology } from '~/react/hooks/useTerminology';

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
        <Text size="large">General</Text>
      </Flex>
      <Flex direction="column" gap={9} style={styles.container}>
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
  const t = useTerminology();
  const navigate = useNavigate({ from: '/' });
  return (
    <>
      <Flex direction="column" gap={5}>
        {isLoading ? (
          <Skeleton height={'16px'} width={'50%'} />
        ) : (
          <Text size={3} variant="secondary">
            If you want to permanently delete this{' '}
            {t.organization({ case: 'lower' })} and all of its data.
          </Text>
        )}
        {isLoading ? (
          <Skeleton height={'32px'} width={'64px'} />
        ) : (
          <Tooltip disabled={canDelete} message={AuthTooltipMessage}>
            <Button
              variant="solid"
              color="danger"
              type="submit"
              onClick={() => navigate({ to: '/delete' })}
              disabled={!canDelete}
              data-test-id="frontier-sdk-delete-organization-btn"
            >
              Delete {t.organization({ case: 'lower' })}
            </Button>
          </Tooltip>
        )}
        <Outlet />
      </Flex>
    </>
  );
};
