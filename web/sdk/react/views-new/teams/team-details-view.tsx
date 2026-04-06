'use client';

import { useCallback, useEffect, useMemo } from 'react';
import {
  DotsHorizontalIcon,
  ExclamationTriangleIcon,
  Pencil1Icon,
  UpdateIcon
} from '@radix-ui/react-icons';
import {
  Breadcrumb,
  Skeleton,
  Flex,
  EmptyState,
  DataTable,
  Menu,
  AlertDialog,
  Dialog,
  IconButton,
  Image
} from '@raystack/apsara-v1';
import deleteIcon from '../../assets/delete.svg';
import { toastManager } from '@raystack/apsara-v1';
import {
  useQuery,
  useMutation,
  createConnectQueryKey,
  useTransport
} from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import {
  FrontierServiceQueries,
  GetGroupRequestSchema,
  ListGroupUsersRequestSchema,
  ListRolesRequestSchema,
  CreatePolicyRequestSchema,
  DeletePolicyRequestSchema,
  ListPoliciesRequestSchema,
  type Role as ProtoRole
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { useFrontier } from '../../contexts/FrontierContext';
import { usePermissions } from '../../hooks/usePermissions';
import { PERMISSIONS, shouldShowComponent } from '../../../utils';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import {
  EditTeamDialog,
  type EditTeamPayload
} from './components/edit-team-dialog';
import {
  DeleteTeamDialog,
  type DeleteTeamPayload
} from './components/delete-team-dialog';
import {
  RemoveMemberDialog,
  type RemoveMemberPayload
} from './components/remove-member-dialog';
import {
  getColumns,
  type MemberMenuPayload
} from './components/member-columns';
import { AddMemberMenu } from './components/add-member-menu';
import styles from './team-details-view.module.css';

const memberMenuHandle = Menu.createHandle<MemberMenuPayload>();
const editTeamDialogHandle = Dialog.createHandle<EditTeamPayload>();
const deleteTeamDialogHandle = AlertDialog.createHandle<DeleteTeamPayload>();
const removeMemberDialogHandle =
  AlertDialog.createHandle<RemoveMemberPayload>();

export interface TeamDetailsViewProps {
  teamId: string;
  teamsLabel?: string;
  onNavigateToTeams?: () => void;
  onDeleteSuccess?: () => void;
}

export function TeamDetailsView({
  teamId,
  teamsLabel = 'Teams',
  onNavigateToTeams,
  onDeleteSuccess
}: TeamDetailsViewProps) {
  const { activeOrganization: organization } = useFrontier();

  const {
    data: teamData,
    isLoading: isTeamLoading,
    error: teamError,
    refetch: refetchTeam
  } = useQuery(
    FrontierServiceQueries.getGroup,
    create(GetGroupRequestSchema, {
      id: teamId || '',
      orgId: organization?.id || ''
    }),
    {
      enabled: !!organization?.id && !!teamId,
      select: d => d
    }
  );

  const team = teamData?.group;

  useEffect(() => {
    if (teamError) {
      toastManager.add({
        title: 'Something went wrong',
        description: teamError.message,
        type: 'error'
      });
    }
  }, [teamError]);

  const {
    data: membersData,
    isLoading: isMembersLoading,
    refetch: refetchMembers
  } = useQuery(
    FrontierServiceQueries.listGroupUsers,
    create(ListGroupUsersRequestSchema, {
      id: teamId || '',
      orgId: organization?.id || '',
      withRoles: true
    }),
    { enabled: !!organization?.id && !!teamId }
  );

  const members = useMemo(() => membersData?.users ?? [], [membersData]);
  const memberRoles = useMemo(() => {
    if (!membersData?.rolePairs) return {};
    return membersData.rolePairs.reduce(
      (acc: Record<string, ProtoRole[]>, mr: { userId: string; roles: ProtoRole[] }) => {
        if (mr.userId) acc[mr.userId] = mr.roles;
        return acc;
      },
      {}
    );
  }, [membersData?.rolePairs]);

  const {
    data: rolesData,
    isLoading: isRolesLoading,
    error: rolesError
  } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, {
      state: 'enabled',
      scopes: [PERMISSIONS.GroupNamespace]
    }),
    { enabled: !!organization?.id && !!teamId }
  );

  const roles = useMemo(() => rolesData?.roles ?? [], [rolesData]);

  useEffect(() => {
    if (rolesError) {
      toastManager.add({
        title: 'Something went wrong',
        description: rolesError.message,
        type: 'error'
      });
    }
  }, [rolesError]);

  const resource = `app/group:${teamId}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      { permission: PERMISSIONS.UpdatePermission, resource },
      { permission: PERMISSIONS.DeletePermission, resource }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!teamId
  );

  const { canUpdateGroup, canDeleteGroup } = useMemo(() => {
    return {
      canUpdateGroup: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      ),
      canDeleteGroup: shouldShowComponent(
        permissions,
        `${PERMISSIONS.DeletePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const isLoading =
    !organization?.id ||
    isTeamLoading ||
    isMembersLoading ||
    isRolesLoading ||
    isPermissionsFetching;

  const refetchAll = useCallback(() => {
    refetchMembers();
  }, [refetchMembers]);

  const columns = useMemo(
    () =>
      getColumns({
        memberRoles,
        roles,
        canUpdateGroup,
        menuHandle: memberMenuHandle
      }),
    [memberRoles, roles, canUpdateGroup]
  );

  const queryClient = useQueryClient();
  const transport = useTransport();

  const { mutateAsync: deletePolicy } = useMutation(
    FrontierServiceQueries.deletePolicy
  );
  const { mutateAsync: createPolicy } = useMutation(
    FrontierServiceQueries.createPolicy
  );

  const updateMemberRole = useCallback(
    async (memberId: string, role: ProtoRole) => {
      try {
        const principal = `${PERMISSIONS.UserNamespace}:${memberId}`;

        const input = create(ListPoliciesRequestSchema, {
          groupId: teamId,
          userId: memberId
        });

        const policiesData = await queryClient.fetchQuery({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.listPolicies,
            transport,
            input,
            cardinality: 'finite'
          })
        });

        const policies =
          (policiesData as { policies?: { id?: string }[] })?.policies ?? [];

        await Promise.all(
          policies.map(p =>
            deletePolicy(
              create(DeletePolicyRequestSchema, { id: p.id || '' })
            )
          )
        );

        await createPolicy(
          create(CreatePolicyRequestSchema, {
            body: {
              roleId: role.id as string,
              resource: `app/group:${teamId}`,
              principal
            }
          })
        );
        refetchAll();
        toastManager.add({
          title: 'Member role updated',
          type: 'success'
        });
      } catch (error) {
        toastManager.add({
          title: 'Something went wrong',
          description:
            error instanceof Error
              ? error.message
              : 'Failed to update member role',
          type: 'error'
        });
      }
    },
    [queryClient, transport, deletePolicy, createPolicy, teamId, refetchAll]
  );

  const handleDeleteSuccess = useCallback(() => {
    onDeleteSuccess?.();
  }, [onDeleteSuccess]);

  const teamTitle = team?.title || '';
  const teamName = team?.name || '';

  return (
    <ViewContainer>
      <ViewHeader
        title={isTeamLoading ? '' : teamTitle}
        breadcrumb={
          <Breadcrumb size="small">
            <Breadcrumb.Item
              href="#"
              onClick={(e) => {
                e.preventDefault();
                onNavigateToTeams?.();
              }}
            >
              {teamsLabel}
            </Breadcrumb.Item>
            <Breadcrumb.Separator />
            <Breadcrumb.Item current>
              {isTeamLoading ? (
                <Skeleton height="16px" width="100px" />
              ) : (
                teamTitle
              )}
            </Breadcrumb.Item>
          </Breadcrumb>
        }
      >
        {!isLoading && (canUpdateGroup || canDeleteGroup) && (
          <TeamActionsMenu
            teamId={teamId}
            teamTitle={teamTitle}
            teamName={teamName}
            canUpdate={canUpdateGroup}
            canDelete={canDeleteGroup}
          />
        )}
      </ViewHeader>

      <DataTable
        data={members}
        columns={columns}
        isLoading={isLoading}
        defaultSort={{ name: 'title', order: 'asc' }}
        mode="client"
      >
        <Flex direction="column" gap={7}>
          <Flex justify="between" gap={3}>
            <Flex gap={3} align="center">
              {isLoading ? (
                <Skeleton height="34px" width="360px" />
              ) : (
                <DataTable.Search
                  placeholder="Search by name or email"
                  size="large"
                  width={360}
                />
              )}
            </Flex>
            {isLoading ? (
              <Skeleton height="34px" width="120px" />
            ) : (
              <AddMemberMenu
                teamId={teamId}
                canUpdateGroup={canUpdateGroup}
                members={members}
                refetch={refetchAll}
              />
            )}
          </Flex>
          <DataTable.Content
            emptyState={
              <EmptyState
                icon={<ExclamationTriangleIcon />}
                heading="No members found"
                subHeading="Get started by adding your first team member."
              />
            }
            classNames={{
              root: styles.tableRoot
            }}
          />
        </Flex>
      </DataTable>

      <Menu handle={memberMenuHandle} modal={false}>
        {({ payload: rawPayload }) => {
          const payload = rawPayload as MemberMenuPayload | undefined;
          return (
            <Menu.Content align="end" className={styles.menuContent}>
              {payload?.excludedRoles.map((role: ProtoRole) => (
                <Menu.Item
                  key={role.id}
                  leadingIcon={<UpdateIcon />}
                  onClick={() =>
                    payload &&
                    updateMemberRole(payload.memberId, role)
                  }
                  data-test-id="frontier-sdk-update-team-member-role-btn"
                >
                  Make {role.title}
                </Menu.Item>
              ))}
              <Menu.Item
                leadingIcon={
                  <Image
                    src={deleteIcon as unknown as string}
                    alt="Remove"
                    width={16}
                    height={16}
                  />
                }
                onClick={() =>
                  payload &&
                  removeMemberDialogHandle.openWithPayload({
                    memberId: payload.memberId,
                    teamId
                  })
                }
                data-test-id="frontier-sdk-remove-team-member-btn"
                style={{
                  color: 'var(--rs-color-foreground-danger-primary)'
                }}
              >
                Remove from team
              </Menu.Item>
            </Menu.Content>
          );
        }}
      </Menu>

      <EditTeamDialog
        handle={editTeamDialogHandle}
        refetch={() => {
          refetchTeam();
        }}
      />
      <DeleteTeamDialog
        handle={deleteTeamDialogHandle}
        refetch={handleDeleteSuccess}
      />
      <RemoveMemberDialog
        handle={removeMemberDialogHandle}
        refetch={refetchAll}
      />
    </ViewContainer>
  );
}

interface TeamActionsMenuProps {
  teamId: string;
  teamTitle: string;
  teamName: string;
  canUpdate: boolean;
  canDelete: boolean;
}

const teamActionsMenuHandle = Menu.createHandle();

function TeamActionsMenu({
  teamId,
  teamTitle,
  teamName,
  canUpdate,
  canDelete
}: TeamActionsMenuProps) {
  return (
    <>
      <Menu.Trigger
        handle={teamActionsMenuHandle}
        render={
          <IconButton
            size={2}
            aria-label="Team actions"
            data-test-id="frontier-sdk-team-details-actions-btn"
          />
        }
      >
        <DotsHorizontalIcon />
      </Menu.Trigger>
      <Menu handle={teamActionsMenuHandle} modal={false}>
        <Menu.Content align="start" className={styles.menuContent}>
          {canUpdate && (
            <Menu.Item
              leadingIcon={<Pencil1Icon />}
              onClick={() =>
                editTeamDialogHandle.openWithPayload({
                  teamId,
                  title: teamTitle,
                  name: teamName
                })
              }
              data-test-id="frontier-sdk-edit-team-details-btn"
            >
              Edit
            </Menu.Item>
          )}
          {canDelete && (
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
                deleteTeamDialogHandle.openWithPayload({ teamId })
              }
              data-test-id="frontier-sdk-delete-team-details-btn"
              style={{
                color: 'var(--rs-color-foreground-danger-primary)'
              }}
            >
              Delete team
            </Menu.Item>
          )}
        </Menu.Content>
      </Menu>
    </>
  );
}
