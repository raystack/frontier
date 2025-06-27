import {
  Flex,
  EmptyState,
  Button,
  Text,
  Skeleton,
  Image,
  DataTable
} from '@raystack/apsara/v1';
import styles from './styles.module.css';
import keyIcon from '~/react/assets/key.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import {
  DEFAULT_API_PLATFORM_APP_NAME,
  DEFAULT_DATE_FORMAT
} from '~/react/utils/constants';
import type { FrontierClientAPIPlatformOptions } from '~/shared/types';
import { useEffect, useMemo, useState } from 'react';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { usePermissions } from '~/react/hooks/usePermissions';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { getColumns } from './columns';
import type { V1Beta1ServiceUser } from '~/api-client';
import { Outlet, useLocation, useNavigate } from '@tanstack/react-router';

const NoServiceAccounts = ({
  config
}: {
  config?: FrontierClientAPIPlatformOptions;
}) => {
  const appName = config?.appName || DEFAULT_API_PLATFORM_APP_NAME;

  const navigate = useNavigate({ from: '/api-keys' });
  return (
    <Flex justify="center" align="center" className={styles.stateContent}>
      <EmptyState
        icon={<Image src={keyIcon as unknown as string} alt="keyIcon" />}
        heading="No service account found"
        subHeading={`Create a new account to use the APIs of ${appName} platform`}
        primaryAction={
          <Button
            data-test-id="frontier-sdk-new-service-account-btn"
            variant="outline"
            color="neutral"
            onClick={() => navigate({ to: '/api-keys/add' })}
          >
            Create new service account
          </Button>
        }
      />
    </Flex>
  );
};

const NoAccess = () => {
  return (
    <Flex justify="center" align="center" className={styles.stateContent}>
      <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="Restricted Access"
        subHeading='Admin access required, please reach out to your admin incase you want to generate a key'
      />
    </Flex>
  );
};

const Headings = ({
  config,
  isLoading
}: {
  config?: FrontierClientAPIPlatformOptions;
  isLoading: boolean;
}) => {
  const appName = config?.appName || DEFAULT_API_PLATFORM_APP_NAME;
  return (
    <Flex direction="column" gap={3} style={{ width: '100%' }}>
      {isLoading ? (
        <Skeleton containerClassName={styles.flex1} />
      ) : (
        <Text size="large">Service Accounts</Text>
      )}
      {isLoading ? (
        <Skeleton containerClassName={styles.flex1} />
      ) : (
        <Text size="regular" variant="secondary">
          Create a non-human identity to allow access to {appName.toLowerCase()}{' '}
          resources
        </Text>
      )}
    </Flex>
  );
};

const useAccess = (orgId?: string) => {
  const resource = `app/organization:${orgId}`;
  const listOfPermissionsToCheck = useMemo(() => {
    return [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource: resource
      }
    ];
  }, [resource]);

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!orgId
  );

  const canUpdateWorkspace = useMemo(() => {
    return shouldShowComponent(
      permissions,
      `${PERMISSIONS.UpdatePermission}::${resource}`
    );
  }, [permissions, resource]);

  return {
    isPermissionsFetching,
    canUpdateWorkspace
  };
};

const ServiceAccountsTable = ({
  isLoading,
  serviceUsers,
  dateFormat
}: {
  isLoading: boolean;
  serviceUsers: V1Beta1ServiceUser[];
  dateFormat?: string;
}) => {
  const columns = getColumns({ dateFormat: dateFormat || DEFAULT_DATE_FORMAT });

  const navigate = useNavigate({ from: '/api-keys' });

  return (
    <DataTable
      data={serviceUsers}
      columns={columns}
      isLoading={isLoading}
      defaultSort={{ name: 'created_at', order: 'desc' }}
      mode="client"
    >
      <Flex direction="column" gap={7} className={styles.tableWrapper}>
        <Flex justify="between" gap={3}>
          <Flex gap={3} justify="start" className={styles.tableSearchWrapper}>
            {isLoading ? (
              <Skeleton height={'32px'} containerClassName={styles.flex1} />
            ) : (
              <DataTable.Search placeholder="Search..." size="medium" />
            )}
          </Flex>
          {isLoading ? (
            <Skeleton height={'32px'} width={'64px'} />
          ) : (
            <Button
              variant="solid"
              color="accent"
              data-test-id="frontier-sdk-add-service-account-btn"
              onClick={() => navigate({ to: '/api-keys/add' })}
            >
              Create
            </Button>
          )}
        </Flex>
        <DataTable.Content
          classNames={{
            root: styles.tableRoot,
            header: styles.tableHeader
          }}
        />
      </Flex>
    </DataTable>
  );
};

export default function ApiKeys() {
  const [serviceUsers, setServiceUsers] = useState<V1Beta1ServiceUser[]>([]);
  const [isServiceUsersLoading, setIsServiceUsersLoading] = useState(false);
  const location = useLocation();
  const refetch = location?.state?.refetch;

  const {
    activeOrganization: organization,
    isActiveOrganizationLoading,
    config,
    client
  } = useFrontier();

  const { isPermissionsFetching, canUpdateWorkspace } = useAccess(
    organization?.id
  );

  useEffect(() => {
    async function getServiceAccounts(orgId: string) {
      try {
        setIsServiceUsersLoading(true);
        const resp = await client?.frontierServiceListOrganizationServiceUsers(
          orgId
        );
        const data = resp?.data?.serviceusers || [];
        setServiceUsers(data);
      } catch (err) {
        console.error(err);
      } finally {
        setIsServiceUsersLoading(false);
      }
    }

    if (organization?.id && canUpdateWorkspace) {
      getServiceAccounts(organization?.id);
    }
  }, [organization?.id, client, canUpdateWorkspace, refetch]);

  const isLoading =
    isActiveOrganizationLoading ||
    isPermissionsFetching ||
    isServiceUsersLoading;

  const serviceAccountsCount: number = serviceUsers?.length;

  return (
    <Flex direction="column" className={styles.container}>
      <Flex className={styles.header}>
        <Text size="large">API</Text>
      </Flex>
      {canUpdateWorkspace || isLoading ? (
        serviceAccountsCount > 0 || isLoading ? (
          <Flex className={styles.content} direction="column" gap={9}>
            <Headings isLoading={isLoading} config={config?.apiPlatform} />
            <ServiceAccountsTable
              isLoading={isLoading}
              serviceUsers={serviceUsers}
              dateFormat={config?.dateFormat}
            />
          </Flex>
        ) : (
          <NoServiceAccounts config={config?.apiPlatform} />
        )
      ) : (
        <NoAccess />
      )}
      <Outlet />
    </Flex>
  );
}
