'use client';

import { useMemo, useCallback, MouseEvent } from 'react';
import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import {
  Breadcrumb,
  Skeleton,
  Flex,
  Text,
  Button,
  Menu,
  AlertDialog,
  Dialog,
  IconButton,
  Image,
  CopyButton
} from '@raystack/apsara-v1';
import deleteIcon from '~/react/assets/delete.svg';
import keyIcon from '~/react/assets/key.svg';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  GetServiceUserRequestSchema,
  type ServiceUserToken
} from '@raystack/proton/frontier';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useTerminology } from '~/react/hooks/useTerminology';
import { usePermissions } from '~/react/hooks/usePermissions';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import { useServiceUserTokens } from './hooks/useServiceUserTokens';
import { ViewContainer } from '~/react/components/view-container';
import { ViewHeader } from '~/react/components/view-header';
import { AddTokenForm } from './components/add-token-form';
import {
  RevokeTokenDialog,
  type RevokeTokenPayload
} from './components/revoke-token-dialog';
import {
  DeleteServiceAccountDialog,
  type DeleteServiceAccountPayload
} from './components/delete-service-account-dialog';
import { ManageProjectAccessDialog } from './components/manage-project-access-dialog';
import styles from './service-account-details-view.module.css';

const actionsMenuHandle = Menu.createHandle();
const revokeTokenDialogHandle = AlertDialog.createHandle<RevokeTokenPayload>();
const deleteDialogHandle = AlertDialog.createHandle<DeleteServiceAccountPayload>();
const manageAccessDialogHandle = Dialog.createHandle();

export interface ServiceAccountDetailsViewProps {
  serviceAccountId: string;
  serviceAccountsLabel?: string;
  onNavigateToServiceAccounts?: () => void;
  onDeleteSuccess?: () => void;
}

export function ServiceAccountDetailsView({
  serviceAccountId,
  serviceAccountsLabel = 'Service accounts',
  onNavigateToServiceAccounts,
  onDeleteSuccess
}: ServiceAccountDetailsViewProps) {
  const { activeOrganization: organization } = useFrontier();
  const t = useTerminology();
  const orgId = organization?.id || '';

  const { data: serviceUser, isLoading: isServiceUserLoading } = useQuery(
    FrontierServiceQueries.getServiceUser,
    create(GetServiceUserRequestSchema, {
      id: serviceAccountId,
      orgId
    }),
    {
      enabled: Boolean(serviceAccountId) && Boolean(orgId),
      select: data => data?.serviceuser
    }
  );

  const {
    tokens: serviceUserTokens,
    isLoading: isTokensLoading,
    addToken,
    removeToken,
    clearFreshTokens
  } = useServiceUserTokens({
    id: serviceAccountId,
    orgId,
    enableFetch: true
  });

  const resource = `app/organization:${orgId}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      { permission: PERMISSIONS.UpdatePermission, resource }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!orgId
  );

  const canUpdateWorkspace = useMemo(
    () =>
      shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      ),
    [permissions, resource]
  );

  const isLoading = !organization?.id || isServiceUserLoading || isTokensLoading || isPermissionsFetching;

  const serviceAccountTitle = serviceUser?.title || '';

  const handleNavigateToServiceAccounts = useCallback(() => {
    clearFreshTokens();
    onNavigateToServiceAccounts?.();
  }, [clearFreshTokens, onNavigateToServiceAccounts]);

  const handleDeleteSuccess = useCallback(() => {
    clearFreshTokens();
    onDeleteSuccess?.();
  }, [clearFreshTokens, onDeleteSuccess]);

  return (
    <ViewContainer>
      <ViewHeader
        title={isLoading ? '' : serviceAccountTitle}
        breadcrumb={
          <Breadcrumb size="small">
            <Breadcrumb.Item
              href="#"
              onClick={(e: MouseEvent) => {
                e.preventDefault();
                handleNavigateToServiceAccounts();
              }}
            >
              {serviceAccountsLabel}
            </Breadcrumb.Item>
            <Breadcrumb.Separator />
            <Breadcrumb.Item current>
              {isLoading ? (
                <Skeleton height="16px" width="100px" />
              ) : (
                serviceAccountTitle
              )}
            </Breadcrumb.Item>
          </Breadcrumb>
        }
      >
        {!isLoading && canUpdateWorkspace && (
          <ActionsMenu serviceAccountId={serviceAccountId} />
        )}
      </ViewHeader>

      <Flex direction="column" gap={7}>
        {isLoading ? (
          <>
            <Skeleton height="20px" width="300px" />
            <Skeleton height="20px" />
          </>
        ) : (
          <>
            <Text size="regular" variant="secondary">
              Create API key for accessing {t.appName()} and its features
            </Text>
            <AddTokenForm
              serviceUserId={serviceAccountId}
              onAddToken={addToken}
            />
          </>
        )}

        {serviceUserTokens.length > 0 && (
          <TokenList
            tokens={serviceUserTokens}
            isLoading={isLoading}
          />
        )}
      </Flex>

      <RevokeTokenDialog
        handle={revokeTokenDialogHandle}
        serviceUserId={serviceAccountId}
        onRevoked={removeToken}
      />
      <DeleteServiceAccountDialog
        handle={deleteDialogHandle}
        refetch={handleDeleteSuccess}
      />
      <ManageProjectAccessDialog
        handle={manageAccessDialogHandle}
        serviceUserId={serviceAccountId}
      />
    </ViewContainer>
  );
}

interface ActionsMenuProps {
  serviceAccountId: string;
}

function ActionsMenu({ serviceAccountId }: ActionsMenuProps) {
  return (
    <>
      <Menu.Trigger
        handle={actionsMenuHandle}
        render={
          <IconButton
            size={2}
            aria-label="Service account actions"
            data-test-id="frontier-sdk-service-account-details-actions-btn"
          />
        }
      >
        <DotsHorizontalIcon />
      </Menu.Trigger>
      <Menu handle={actionsMenuHandle} modal={false}>
        <Menu.Content align="start" className={styles.menuContent}>
          <Menu.Item
            leadingIcon={
              <Image
                src={keyIcon as unknown as string}
                alt="Manage access"
                width={16}
                height={16}
              />
            }
            onClick={() => manageAccessDialogHandle.open(null)}
            data-test-id="frontier-sdk-service-account-manage-access-btn"
          >
            Manage access
          </Menu.Item>
          <Menu.Item
            leadingIcon={
              <Image
                src={deleteIcon as unknown as string}
                alt="Delete"
                width={16}
                height={16}
              />
            }
            onClick={() =>
              deleteDialogHandle.openWithPayload({
                serviceAccountId
              })
            }
            data-test-id="frontier-sdk-service-account-delete-btn"
            style={{ color: 'var(--rs-color-foreground-danger-primary)' }}
          >
            Delete account
          </Menu.Item>
        </Menu.Content>
      </Menu>
    </>
  );
}

function TokenList({
  tokens,
  isLoading
}: {
  tokens: ServiceUserToken[];
  isLoading: boolean;
}) {
  if (isLoading) {
    return (
      <Flex direction="column" gap={3}>
        <Skeleton height="64px" />
        <Skeleton height="64px" />
      </Flex>
    );
  }

  return (
    <div className={styles.tokenList}>
      {tokens.map(token => (
        <TokenItem key={token.id} token={token} />
      ))}
    </div>
  );
}

function TokenItem({ token }: { token: ServiceUserToken }) {
  const encodedToken = 'Basic ' + btoa(`${token?.id}:${token?.token}`);

  return (
    <Flex className={styles.tokenItem} direction="column" gap={3}>
      <Flex justify="between" align="center">
        <Text size="regular" weight="medium">
          {token?.title}
        </Text>
        <Button
          variant="outline"
          color="neutral"
          size="small"
          onClick={() =>
            revokeTokenDialogHandle.openWithPayload({ tokenId: token?.id || '' })
          }
          data-test-id="frontier-sdk-service-account-token-revoke-btn"
        >
          Revoke
        </Button>
      </Flex>
      {token?.token ? (
        <>
          <Text size="small" variant="secondary">
            Note: Please save your key securely, it cannot be recovered after
            leaving this page
          </Text>
          <Flex align="center" gap={3}>
            <Text size="regular" weight="medium" className={styles.tokenText}>
              {encodedToken}
            </Text>
            <CopyButton
              text={encodedToken}
              size={2}
              className={styles.copyButton}
              data-test-id="frontier-sdk-service-account-token-copy-btn"
            />
          </Flex>
        </>
      ) : null}
    </Flex>
  );
}
