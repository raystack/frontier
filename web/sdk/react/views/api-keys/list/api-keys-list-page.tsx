import {
  Flex,
  EmptyState,
  Button,
  Skeleton,
  Image,
  DataTable
} from '@raystack/apsara';
import keyIcon from '~/react/assets/key.svg';
import { PageHeader } from '~/react/components/common/page-header';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import { useMemo, useState } from 'react';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { usePermissions } from '~/react/hooks/usePermissions';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { getColumns } from './api-keys-columns';
import { useTerminology } from '~/react/hooks/useTerminology';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  ListOrganizationServiceUsersRequestSchema,
  type ServiceUser
} from '@raystack/proton/frontier';

import sharedStyles from '../../../components/organization/styles.module.css';
import styles from './api-keys.module.css';
import { AddServiceAccountDialog } from './add-service-account-dialog';
import { DeleteServiceAccountDialog } from './delete-service-account-dialog';

const NoServiceAccounts = ({
  onCreateClick
}: {
  onCreateClick: () => void;
}) => {
  const t = useTerminology();

  return (
    <Flex justify="center" align="center" className={styles.stateContent}>
      <EmptyState
        icon={<Image src={keyIcon as unknown as string} alt="keyIcon" />}
        heading="No service account found"
        subHeading={`Create a new account to use the APIs of ${t.appName()} platform`}
        primaryAction={
          <Button
            data-test-id="frontier-sdk-new-service-account-btn"
            variant="outline"
            color="neutral"
            onClick={onCreateClick}
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
        subHeading="Admin access required, please reach out to your admin incase you want to generate a key"
      />
    </Flex>
  );
};

interface ApiKeysHeaderProps {
  isLoading?: boolean;
}

const ApiKeysHeader = ({ isLoading }: ApiKeysHeaderProps) => {
  const t = useTerminology();

  if (isLoading) {
    return (
      <Flex direction="column" gap={2} className={styles.flex1}>
        <Skeleton />
        <Skeleton />
      </Flex>
    );
  }

  return (
    <PageHeader
      title="API Keys"
      description={`Create a non-human identity to allow access to ${t.appName()} resources.`}
    />
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

interface ServiceAccountsTableProps {
  isLoading: boolean;
  serviceUsers: ServiceUser[];
  dateFormat?: string;
  onServiceAccountClick?: (id: string) => void;
  onCreateClick: () => void;
  onDeleteClick: (id: string) => void;
}

const ServiceAccountsTable = ({
  isLoading,
  serviceUsers,
  dateFormat,
  onServiceAccountClick,
  onCreateClick,
  onDeleteClick
}: ServiceAccountsTableProps) => {
  const columns = getColumns({
    dateFormat: dateFormat || DEFAULT_DATE_FORMAT,
    onDeleteClick
  });

  return (
    <DataTable
      data={serviceUsers}
      columns={columns}
      isLoading={isLoading}
      defaultSort={{ name: 'createdAt', order: 'desc' }}
      mode="client"
      onRowClick={row => onServiceAccountClick?.(row.id)}
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
              onClick={onCreateClick}
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

export interface ApiKeysListPageProps {
  onServiceAccountClick?: (id: string) => void;
}

export default function ApiKeysListPage({
  onServiceAccountClick
}: ApiKeysListPageProps) {
  const {
    activeOrganization: organization,
    isActiveOrganizationLoading,
    config
  } = useFrontier();

  const { isPermissionsFetching, canUpdateWorkspace } = useAccess(
    organization?.id
  );

  const {
    data: serviceUsersData,
    isLoading: isServiceUsersLoading
  } = useQuery(
    FrontierServiceQueries.listOrganizationServiceUsers,
    create(ListOrganizationServiceUsersRequestSchema, {
      id: organization?.id ?? ''
    }),
    {
      enabled: Boolean(organization?.id) && canUpdateWorkspace
    }
  );

  const serviceUsers = useMemo(() => serviceUsersData?.serviceusers ?? [], [serviceUsersData]);

  const isLoading =
    isActiveOrganizationLoading ||
    isPermissionsFetching ||
    isServiceUsersLoading;

  const serviceAccountsCount: number = serviceUsers?.length;

  const [showAddDialog, setShowAddDialog] = useState(false);

  const [deleteState, setDeleteState] = useState({
    open: false,
    serviceAccountId: ''
  });

  const handleDeleteOpenChange = (value: boolean) => {
    if (!value) {
      setDeleteState({ open: false, serviceAccountId: '' });
    } else {
      setDeleteState(prev => ({ ...prev, open: value }));
    }
  };

  const handleDeleteClick = (id: string) => {
    setDeleteState({ open: true, serviceAccountId: id });
  };

  const handleAddOpenChange = (value: boolean) => {
    setShowAddDialog(value);
  };

  const handleCreated = (serviceUserId: string) => {
    setShowAddDialog(false);
    onServiceAccountClick?.(serviceUserId);
  };

  return (
    <Flex direction="column" className={sharedStyles.pageWrapper}>
      <Flex direction="column" className={`${sharedStyles.container} ${sharedStyles.containerFlex}`}>
        <Flex direction="row" justify="between" align="center" className={sharedStyles.header}>
          <ApiKeysHeader isLoading={isLoading} />
        </Flex>
        {canUpdateWorkspace || isLoading ? (
          serviceAccountsCount > 0 || isLoading ? (
            <Flex direction="column" gap={9} className={sharedStyles.contentWrapper}>
              <ServiceAccountsTable
                isLoading={isLoading}
                serviceUsers={serviceUsers}
                dateFormat={config?.dateFormat}
                onServiceAccountClick={onServiceAccountClick}
                onCreateClick={() => setShowAddDialog(true)}
                onDeleteClick={handleDeleteClick}
              />
            </Flex>
          ) : (
            <NoServiceAccounts onCreateClick={() => setShowAddDialog(true)} />
          )
        ) : (
          <NoAccess />
        )}
      </Flex>
      <AddServiceAccountDialog
        open={showAddDialog}
        onOpenChange={handleAddOpenChange}
        onCreated={handleCreated}
      />
      <DeleteServiceAccountDialog
        open={deleteState.open}
        onOpenChange={handleDeleteOpenChange}
        serviceAccountId={deleteState.serviceAccountId}
      />
    </Flex>
  );
}
