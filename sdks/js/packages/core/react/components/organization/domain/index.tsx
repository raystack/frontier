'use client';

import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import {
  Button,
  Tooltip,
  EmptyState,
  Skeleton,
  Text,
  Flex,
  DataTable
} from '@raystack/apsara';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useEffect, useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationDomains } from '~/react/hooks/useOrganizationDomains';
import { usePermissions } from '~/react/hooks/usePermissions';
import type { V1Beta1Domain } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { getColumns } from './domain.columns';
import { AuthTooltipMessage } from '~/react/utils';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import { PageHeader } from '~/react/components/common/page-header';
import sharedStyles from '../styles.module.css';
import styles from './domain.module.css';

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
      <Flex direction="column" className={styles.container}>
        <Flex direction="row" justify="between" align="center" className={sharedStyles.header}>
          <PageHeader 
            title="Allowed email domains" 
            description="Anyone with an email address at these domains is allowed to sign up
          for this organization."
          />
        </Flex>
        <Flex direction="column" gap={9}>
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
  const navigate = useNavigate({ from: '/domains' });
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
    <DataTable
      data={domains}
      isLoading={isLoading}
      defaultSort={{ name: 'name', order: 'asc' }}
      columns={columns}
      mode="client"
    >
      <Flex direction="column" gap={7} className={styles.tableWrapper}>
        <Flex justify="between" gap={3}>
          <Flex gap={3} justify="start" className={styles.tableSearchWrapper}>
            {isLoading ? (
              <Skeleton height="34px" width="500px" />
            ) : (
              <DataTable.Search placeholder="Search by name " size="large" />
            )}
          </Flex>
          {isLoading ? (
            <Skeleton height="34px" width="64px" />
          ) : (
            <Tooltip
              message={AuthTooltipMessage}
              side="left"
              disabled={canCreateDomain}
            >
              <Button
                disabled={!canCreateDomain}
                size="normal"
                style={{ width: 'fit-content' }}
                onClick={() => navigate({ to: '/domains/modal' })}
                data-test-id="frontier-sdk-add-domain-btn"
              >
                Add Domain
              </Button>
            </Tooltip>
          )}
        </Flex>
        <DataTable.Content
          emptyState={noDataChildren}
          classNames={{
            root: styles.tableRoot,
            header: styles.tableHeader
          }}
        />
      </Flex>
    </DataTable>
  );
};

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="No domains found"
    subHeading="Get started by adding your first domain."
  />
);
