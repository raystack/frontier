'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationDomains } from '~/react/hooks/useOrganizationDomains';
import { usePermissions } from '~/react/hooks/usePermissions';
import { V1Beta1Domain } from '~/src';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { styles } from '../styles';
import { getColumns } from './domain.columns';

export default function Domain() {
  const { isFetching, domains } = useOrganizationDomains();
  const { activeOrganization: organization } = useFrontier();

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = [
    {
      permission: PERMISSIONS.UpdatePermission,
      resource
    }
  ];

  const { permissions } = usePermissions(
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

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Domains</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <AllowedEmailDomains />
          {/* @ts-ignore */}
          <Domains
            domains={domains}
            isLoading={isFetching}
            canCreateDomain={canCreateDomain}
          />
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}

const AllowedEmailDomains = () => {
  let navigate = useNavigate({ from: '/domains' });
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap="small">
        <Text size={6}>Allowed email domains</Text>
        <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
          Anyone with an email address at these domains is allowed to sign up
          for this workspace.
        </Text>
      </Flex>
    </Flex>
  );
};

const Domains = ({
  domains,
  isLoading,
  canCreateDomain
}: {
  domains: V1Beta1Domain[];
  isLoading?: boolean;
  canCreateDomain?: boolean;
}) => {
  let navigate = useNavigate({ from: '/domains' });

  const tableStyle = domains?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const columns = useMemo(
    () => getColumns(canCreateDomain, isLoading),
    [canCreateDomain, isLoading]
  );
  return (
    <Flex direction="row">
      <DataTable
        data={domains ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 212px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar style={{ padding: 0, border: 0 }}>
          <Flex justify="between" gap="small">
            <Flex style={{ maxWidth: '360px', width: '100%' }}>
              <DataTable.GloabalSearch
                placeholder="Search by name"
                size="medium"
              />
            </Flex>

            {canCreateDomain ? (
              <Button
                variant="primary"
                style={{ width: 'fit-content' }}
                onClick={() => navigate({ to: '/domains/modal' })}
              >
                Add Domain
              </Button>
            ) : null}
          </Flex>
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
};

const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 domains in your organization</h3>
    <div className="pera">Try adding new domains.</div>
  </EmptyState>
);
