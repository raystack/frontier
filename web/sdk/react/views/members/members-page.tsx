'use client';

import { useMemo, useState } from 'react';

import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import {
  Button,
  Tooltip,
  Skeleton,
  EmptyState,
  Flex,
  DataTable
} from '@raystack/apsara';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useOrganizationMembers } from '~/react/hooks/useOrganizationMembers';
import { usePermissions } from '~/react/hooks/usePermissions';
import { AuthTooltipMessage } from '~/react/utils';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import type { MembersTableType } from './member-types';
import { PageHeader } from '~/react/components/common/page-header';
import { InviteMemberDialog } from './invite-member-dialog';
import { RemoveMemberDialog } from './remove-member-dialog';
import sharedStyles from '../../components/organization/styles.module.css';
import styles from './members.module.css';
import { getColumns } from './member-columns';

export interface MembersPageProps {
  title?: string;
  description?: string;
}

export function MembersPage({
  title = 'Members',
  description = 'Manage members in this domain.'
}: MembersPageProps = {}) {
  const { activeOrganization: organization } = useFrontier();

  const resource = `app/organization:${organization?.id}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.InvitationCreatePermission,
        resource
      },
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

  const { canCreateInvite, canDeleteUser } = useMemo(() => {
    return {
      canCreateInvite: shouldShowComponent(
        permissions,
        `${PERMISSIONS.InvitationCreatePermission}::${resource}`
      ),
      canDeleteUser: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const {
    roles,
    members,
    memberRoles,
    refetch,
    isFetching: isOrgMembersLoading
  } = useOrganizationMembers({
    showInvitations: canCreateInvite
  });

  const isLoading = isOrgMembersLoading || isPermissionsFetching;

  const [showInviteDialog, setShowInviteDialog] = useState(false);
  const [removeMemberState, setRemoveMemberState] = useState<{
    open: boolean;
    memberId: string;
    invited: string;
  }>({ open: false, memberId: '', invited: 'false' });

  const handleRemoveMember = (memberId: string, invited: string) => {
    setRemoveMemberState({ open: true, memberId, invited });
  };

  const handleInviteOpenChange = (value: boolean) => {
    setShowInviteDialog(value);
    refetch();
  };

  const handleRemoveOpenChange = (value: boolean) => {
    setRemoveMemberState({ open: value, memberId: '', invited: 'false' });
    refetch();
  };

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
          <PageHeader title={title} description={description} />
        </Flex>
        <Flex
          direction="column"
          gap={9}
          className={sharedStyles.contentWrapper}
        >
          {organization?.id ? (
            <MembersTable
              roles={roles}
              users={members}
              organizationId={organization?.id}
              isLoading={isLoading}
              canCreateInvite={canCreateInvite}
              canDeleteUser={canDeleteUser}
              memberRoles={memberRoles}
              refetch={refetch}
              onRemoveMember={handleRemoveMember}
              onInviteClick={() => setShowInviteDialog(true)}
            />
          ) : null}
        </Flex>
      </Flex>
      <InviteMemberDialog
        open={showInviteDialog}
        onOpenChange={handleInviteOpenChange}
      />
      <RemoveMemberDialog
        open={removeMemberState.open}
        onOpenChange={handleRemoveOpenChange}
        memberId={removeMemberState.memberId}
        invited={removeMemberState.invited}
      />
    </Flex>
  );
}

const MembersTable = ({
  isLoading,
  users,
  canCreateInvite,
  canDeleteUser,
  organizationId,
  memberRoles,
  roles,
  refetch,
  onRemoveMember,
  onInviteClick
}: MembersTableType & { onInviteClick: () => void }) => {
  const columns = useMemo(
    () =>
      getColumns(
        organizationId,
        memberRoles,
        roles,
        canDeleteUser,
        refetch,
        onRemoveMember
      ),
    [organizationId, memberRoles, canDeleteUser, roles, refetch, onRemoveMember]
  );

  return (
    <DataTable
      data={users}
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
              <DataTable.Search placeholder="Search by name or email" />
            )}
          </Flex>
          {isLoading ? (
            <Skeleton height="34px" width="64px" />
          ) : (
            <Tooltip
              message={AuthTooltipMessage}
              side="left"
              disabled={canCreateInvite}
            >
              <Button
                size="small"
                style={{ width: 'fit-content', height: '100%' }}
                onClick={onInviteClick}
                disabled={!canCreateInvite}
                data-test-id="frontier-sdk-remove-member-link"
              >
                Invite people
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
    heading="No members found"
    subHeading="Get started by adding your first member"
  />
);
