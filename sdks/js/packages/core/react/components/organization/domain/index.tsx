'use client';

import {
  DataTable,
} from '@raystack/apsara';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { Button, Tooltip, EmptyState, Skeleton, Text, Flex } from '@raystack/apsara/v1';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useEffect, useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationDomains } from '~/react/hooks/useOrganizationDomains';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1Domain } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { styles } from '../styles';
import { getColumns } from './domain.columns';
import { AuthTooltipMessage } from '~/react/utils';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';

export default function Domain() {
  const { isFetching, domains, refetch } = useOrganizationDomains();
  const { activeOrganization: organization, config } = useFrontier();

  const routerState = useRouterState();

  const isListRoute = useMemo(() => {
    return routerState.location.pathname === '/domains';
  }, [routerState.location.pathname]);

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const { canCreateDomain } = useMemo(() => {
    return {
      canCreateDomain: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  useEffect(() => {
    if (isListRoute) {
      refetch();
    }
  }, [isListRoute, refetch, routerState.location.state.key]);

  const isLoading = isFetching || isPermissionsFetching;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size="large">Security</Text>
      </Flex>
      <Flex direction="column" gap={9} style={styles.container}>
        <Flex direction="column" gap={7}>
          <AllowedEmailDomains />
          <Domains
            domains={domains}
            isLoading={isLoading}
            canCreateDomain={canCreateDomain}
            dateFormat={config?.dateFormat}
          />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}

const AllowedEmailDomains = () => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap={3}>
        <Text size="large" weight="medium">Allowed email domains</Text>
        <Text size="regular" variant="secondary">
          Anyone with an email address at these domains is allowed to sign up
          for this workspace.
        </Text>
      </Flex>
    </Flex>
  );
};

const Domains = ({
  domains = [],
  isLoading,
  canCreateDomain,
  dateFormat
}: {
  domains: V1Beta1Domain[];
  isLoading?: boolean;
  canCreateDomain?: boolean;
  dateFormat?: string;
}) => {
  let navigate = useNavigate({ from: '/domains' });
  const tableStyle = useMemo(
    () =>
      domains?.length ? { width: '100%' } : { width: '100%', height: '100%' },
    [domains?.length]
  );

  const columns = useMemo(
    () =>
      getColumns({
        canCreateDomain,
        dateFormat: dateFormat || DEFAULT_DATE_FORMAT
      }),
    [canCreateDomain, dateFormat]
  );

  return (
    <Flex direction="row">
      <DataTable
        data={domains}
        isLoading={isLoading}
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 212px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar
          style={{ padding: 0, border: 0, marginBottom: 'var(--pd-16)' }}
        >
          <Flex justify="between" gap={3}>
            <Flex style={{ maxWidth: '360px', width: '100%' }}>
              <DataTable.GloabalSearch
                placeholder="Search by name"
                size="medium"
              />
            </Flex>
            {isLoading ? (
              <Skeleton height={'32px'} width={'64px'} />
            ) : (
              <Tooltip
                message={AuthTooltipMessage}
                side="left"
                disabled={canCreateDomain}
              >
                <Button
                  disabled={!canCreateDomain}
                  size="small"
                  style={{ width: 'fit-content' }}
                  onClick={() => navigate({ to: '/domains/modal' })}
                  data-test-id="frontier-sdk-add-domain-btn"
                >
                  Add Domain
                </Button>
              </Tooltip>
            )}
          </Flex>
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
};

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading={"0 domains in your organization"}
    subHeading={"Try adding new domains."}
  />
);
