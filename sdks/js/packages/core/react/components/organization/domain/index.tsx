'use client';

import { Button, DataTable, EmptyState, Flex, Text } from '@raystack/apsara';
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Domain } from '~/src';
import { styles } from '../styles';
import { getColumns } from './domain.columns';

export default function Domain() {
  const { client, activeOrganization: organization } = useFrontier();
  const routerState = useRouterState();
  const [domains, setDomains] = useState([]);
  const [isDomainsLoading, setIsDomainsLoading] = useState(false);

  const getDomains = useCallback(async () => {
    try {
      setIsDomainsLoading(true);
      if (!organization?.id) return;
      const {
        // @ts-ignore
        data: { domains = [] }
      } = await client?.frontierServiceListOrganizationDomains(
        organization?.id
      );
      setDomains(domains);
    } catch (err) {
      console.error(err);
    } finally {
      setIsDomainsLoading(false);
    }
  }, [client, organization?.id]);

  useEffect(() => {
    getDomains();
  }, [getDomains, routerState.location.key]);

  useEffect(() => {
    getDomains();
  }, [client, getDomains, organization?.id]);

  const updatedDomains = useMemo(
    () =>
      isDomainsLoading
        ? [{ id: 1 }, { id: 2 }, { id: 3 }]
        : domains.length
        ? domains
        : [],
    [isDomainsLoading, domains]
  );

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Domains</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <AllowedEmailDomains />
          {/* @ts-ignore */}
          <Domains domains={updatedDomains} isLoading={isDomainsLoading} />
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
  isLoading
}: {
  domains: V1Beta1Domain[];
  isLoading?: boolean;
}) => {
  let navigate = useNavigate({ from: '/domains' });

  const tableStyle = domains?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const columns = useMemo(() => getColumns(isLoading), [isLoading]);
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

            <Button
              variant="primary"
              style={{ width: 'fit-content' }}
              onClick={() => navigate({ to: '/domains/modal' })}
            >
              Add Domain
            </Button>
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
