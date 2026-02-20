'use client';

import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import {
  Button,
  Tooltip,
  EmptyState,
  Skeleton,
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
import { PageHeader } from '~/react/components/common/page-header';
import sharedStyles from '../styles.module.css';
import styles from './domain.module.css';
import { useTerminology } from '~/react/hooks/useTerminology';

export default function Domain() {
  const { isFetching, domains, error: domainsError } = useOrganizationDomains();
  const { activeOrganization: organization, config } = useFrontier();
  const t = useTerminology();

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
    <Flex direction="column" className={sharedStyles.pageWrapper}>
      <Flex
        direction="column"
        className={`${sharedStyles.container} ${sharedStyles.containerFlex}`}
      >
        <Flex
          direction="row"
          justify="between"
          align="center"
          className={sharedStyles.header}
        >
          <PageHeader
            title="Allowed email domains"
            description={`Anyone with an email address at these domains is allowed to sign up
          for this ${t.organization({ case: 'lower' })}.`}
          />
        </Flex>
        <Flex
          direction="column"
          gap={9}
          className={sharedStyles.contentWrapper}
        >
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
