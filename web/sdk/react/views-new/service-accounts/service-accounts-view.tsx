'use client';

import { useMemo, useState } from 'react';
import { ExclamationTriangleIcon, KeyboardIcon, TrashIcon } from '@radix-ui/react-icons';
import {
  Button,
  Tooltip,
  Skeleton,
  Flex,
  EmptyState,
  DataTable,
  Dialog,
  AlertDialog,
  Menu
} from '@raystack/apsara-v1';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  ListOrganizationServiceUsersRequestSchema
} from '@raystack/proton/frontier';
import { useFrontier } from '../../contexts/FrontierContext';
import { usePermissions } from '../../hooks/usePermissions';
import { useTerminology } from '../../hooks/useTerminology';
import { AuthTooltipMessage } from '../../utils';
import { PERMISSIONS, shouldShowComponent } from '../../../utils';
import { DEFAULT_DATE_FORMAT } from '../../utils/constants';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import {
  getColumns,
  type ServiceAccountMenuPayload
} from './components/service-account-columns';
import { AddServiceAccountDialog } from './components/add-service-account-dialog';
import {
  DeleteServiceAccountDialog,
  type DeleteServiceAccountPayload
} from './components/delete-service-account-dialog';
import { ManageProjectAccessDialog } from './components/manage-project-access-dialog';
import styles from './service-accounts-view.module.css';

const serviceAccountMenuHandle = Menu.createHandle<ServiceAccountMenuPayload>();
const addDialogHandle = Dialog.createHandle();
const deleteDialogHandle = AlertDialog.createHandle<DeleteServiceAccountPayload>();
const manageAccessDialogHandle = Dialog.createHandle<string>();

export interface ServiceAccountsViewProps {
  onServiceAccountClick?: (id: string) => void;
}

export function ServiceAccountsView({
  onServiceAccountClick
}: ServiceAccountsViewProps) {
  const {
    activeOrganization: organization,
    isActiveOrganizationLoading,
    config
  } = useFrontier();
  const t = useTerminology();

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

  const canUpdateWorkspace = useMemo(
    () =>
      shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      ),
    [permissions, resource]
  );

  const orgId = organization?.id ?? '';
  const [manageAccessServiceUserId, setManageAccessServiceUserId] = useState('');

  const {
    data: serviceUsersData,
    isLoading: isServiceUsersLoading,
    refetch
  } = useQuery(
    FrontierServiceQueries.listOrganizationServiceUsers,
    create(ListOrganizationServiceUsersRequestSchema, {
      id: orgId
    }),
    {
      enabled: Boolean(orgId) && canUpdateWorkspace
    }
  );

  const serviceUsers = useMemo(
    () => serviceUsersData?.serviceusers ?? [],
    [serviceUsersData]
  );

  const isPermissionsLoading =
    isActiveOrganizationLoading || isPermissionsFetching;

  const isLoading = isPermissionsLoading || isServiceUsersLoading;

  const dateFormat = config?.dateFormat || DEFAULT_DATE_FORMAT;

  const columns = useMemo(
    () =>
      getColumns({
        dateFormat,
        menuHandle: serviceAccountMenuHandle,
        canUpdateWorkspace,
        orgId
      }),
    [dateFormat, canUpdateWorkspace, orgId]
  );

  const handleCreated = (serviceUserId: string) => {
    onServiceAccountClick?.(serviceUserId);
  };

  const handleRefetch = () => {
    refetch();
  };

  const hasNoAccess = !canUpdateWorkspace && !isPermissionsLoading;
  const hasNoServiceAccounts =
    canUpdateWorkspace && !isLoading && serviceUsers.length === 0;

  return (
    <ViewContainer>
      <ViewHeader
        title="Service accounts"
        description={`Create and manage service accounts for secure, automated access to ${t.appName()}.`}
      />

      {hasNoAccess ? (
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="Restricted Access"
          subHeading="Admin access required, please reach out to your admin incase you want to generate a key."
        />
      ) : hasNoServiceAccounts ? (
        <EmptyState
          icon={<KeyboardIcon />}
          heading="No Service Account Found"
          subHeading={`Create a new account to use the APIs of ${t.appName()} platform`}
          primaryAction={
            <Button
              variant="outline"
              color="neutral"
              onClick={() => addDialogHandle.open(null)}
              data-test-id="frontier-sdk-new-service-account-btn"
            >
              Add service account
            </Button>
          }
        />
      ) : (
        <DataTable
          data={serviceUsers}
          columns={columns}
          isLoading={isLoading}
          defaultSort={{ name: 'createdAt', order: 'desc' }}
          mode="client"
          onRowClick={row => onServiceAccountClick?.(row.id)}
        >
          <Flex direction="column" gap={7}>
            <Flex justify="between" gap={3}>
              <Flex gap={3} align="center">
                {isLoading ? (
                  <Skeleton height="34px" width="360px" />
                ) : (
                  <DataTable.Search
                    placeholder="Search by name."
                    size="large"
                    width={360}
                  />
                )}
              </Flex>
              {isLoading ? (
                <Skeleton height="34px" width="160px" />
              ) : (
                <Tooltip>
                  <Tooltip.Trigger
                    disabled={canUpdateWorkspace}
                    render={<span />}
                  >
                    <Button
                      variant="solid"
                      color="accent"
                      onClick={() => addDialogHandle.open(null)}
                      disabled={!canUpdateWorkspace}
                      data-test-id="frontier-sdk-add-service-account-btn"
                    >
                      Add service account
                    </Button>
                  </Tooltip.Trigger>
                  {!canUpdateWorkspace && (
                    <Tooltip.Content>{AuthTooltipMessage}</Tooltip.Content>
                  )}
                </Tooltip>
              )}
            </Flex>
            <DataTable.Content
              emptyState={
                <EmptyState
                  icon={<ExclamationTriangleIcon />}
                  heading="No service accounts found"
                  subHeading="Try adjusting your search"
                />
              }
              classNames={{
                root: styles.tableRoot
              }}
            />
          </Flex>
        </DataTable>
      )}

      <Menu handle={serviceAccountMenuHandle} modal={false}>
        {({ payload: rawPayload }) => {
          const payload = rawPayload as ServiceAccountMenuPayload | undefined;
          return (
            <Menu.Content align="end" className={styles.menuContent}>
              {payload?.canManageAccess && (
                <Menu.Item
                  leadingIcon={<KeyboardIcon />}
                  onClick={() => {
                    if (payload) {
                      setManageAccessServiceUserId(payload.serviceAccountId);
                      manageAccessDialogHandle.open(null);
                    }
                  }}
                  data-test-id="frontier-sdk-manage-access-menu-item"
                >
                  Manage Access
                </Menu.Item>
              )}
              {payload?.canDelete && (
                <Menu.Item
                  leadingIcon={<TrashIcon />}
                  onClick={() =>
                    payload &&
                    deleteDialogHandle.openWithPayload({
                      serviceAccountId: payload.serviceAccountId
                    })
                  }
                  data-test-id="frontier-sdk-delete-account-menu-item"
                >
                  Delete Account
                </Menu.Item>
              )}
            </Menu.Content>
          );
        }}
      </Menu>

      <AddServiceAccountDialog
        handle={addDialogHandle}
        onCreated={handleCreated}
      />
      <DeleteServiceAccountDialog
        handle={deleteDialogHandle}
        refetch={handleRefetch}
      />
      {manageAccessServiceUserId && (
        <ManageProjectAccessDialog
          handle={manageAccessDialogHandle}
          serviceUserId={manageAccessServiceUserId}
        />
      )}
    </ViewContainer>
  );
}
