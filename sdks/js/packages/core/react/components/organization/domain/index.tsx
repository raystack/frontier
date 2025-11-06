'use client';

import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import {
  Button,
  Tooltip,
  EmptyState,
  Skeleton,
  Text,
  Flex,
  DataTable,
  toast
} from '@raystack/apsara';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { useEffect, useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationDomains } from '~/react/hooks/useOrganizationDomains';
import { usePermissions } from '~/react/hooks/usePermissions';
import { type Domain as DomainType } from '@raystack/proton/frontier';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { getColumns } from './domain.columns';
import { AuthTooltipMessage } from '~/react/utils';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import styles from './domain.module.css';

export default function Domain() {
  const { isFetching, domains, error: domainsError } = useOrganizationDomains();
  const { activeOrganization: organization, config } = useFrontier();

  useEffect(() => {
    if (domainsError) {
      toast.error('Something went wrong', {
        description: (domainsError as Error).message
      });
    }
  }, [domainsError]);

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

  const isLoading = isFetching || isPermissionsFetching;

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex className={styles.header}>
        <Text size="large">Security</Text>
      </Flex>
      <Flex direction="column" gap={9} className={styles.container}>
        <AllowedEmailDomains />
        <Domains
          domains={domains}
          isLoading={isLoading}
          canCreateDomain={canCreateDomain}
          dateFormat={config?.dateFormat}
        />
      </Flex>
      <Outlet />
    </Flex>
  );
}

const AllowedEmailDomains = () => {
  return (
    <Flex direction="row" justify="between" align="center">
      <Flex direction="column" gap={3}>
        <Text size="large" weight="medium">
          Allowed email domains
        </Text>
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
  domains: DomainType[];
  isLoading?: boolean;
  canCreateDomain?: boolean;
  dateFormat?: string;
}) => {
  const navigate = useNavigate({ from: '/domains' });

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
