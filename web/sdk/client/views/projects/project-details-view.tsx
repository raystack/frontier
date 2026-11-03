'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
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
  Image,
  Select
} from '@raystack/apsara';
import deleteIcon from '../../assets/delete.svg';
import { toastManager } from '@raystack/apsara';
import { useQuery } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  ListProjectGroupsRequestSchema,
  ListProjectUsersRequestSchema,
  GetProjectRequestSchema,
  ListRolesRequestSchema,
  type Role as ProtoRole
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { useFrontier } from '../../contexts/FrontierContext';
import { usePermissions } from '../../hooks/usePermissions';
import { PERMISSIONS, shouldShowComponent } from '../../../utils';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { DeleteBlockedDialogHost } from '../../components/delete-blocked-dialog';
import type { DeleteBlockedPayload } from '../../../shared/types';
import {
  EditProjectDialog,
  type EditProjectPayload
} from './components/edit-project-dialog';
import {
  DeleteProjectDialog,
  type DeleteProjectPayload
} from './components/delete-project-dialog';
import { DeleteProjectMenuItem } from './components/delete-project-menu-item';
import { useProjectDeleteCheck } from './hooks/useProjectDeleteCheck';
import {
  RemoveMemberDialog,
  type RemoveMemberPayload
} from './components/remove-member-dialog';
import {
  UpdateRoleDialog,
  type UpdateRolePayload
} from './components/update-role-dialog';
import {
  getColumns,
  type MemberRow,
  type MemberMenuPayload
} from './components/member-columns';
import { AddMemberMenu } from './components/add-member-menu';
import styles from './project-details-view.module.css';

interface ProjectGroupRolePair {
  groupId?: string;
  roles: ProtoRole[];
}

interface ProjectUserRolePair {
  userId?: string;
  roles: ProtoRole[];
}

const memberMenuHandle = Menu.createHandle<MemberMenuPayload>();
const editProjectDialogHandle = Dialog.createHandle<EditProjectPayload>();
const removeMemberDialogHandle =
  AlertDialog.createHandle<RemoveMemberPayload>();
const updateRoleDialogHandle = AlertDialog.createHandle<UpdateRolePayload>();

export interface ProjectDetailsViewProps {
  projectId: string;
  projectsLabel?: string;
  onNavigateToProjects?: () => void;
  onDeleteSuccess?: () => void;
}

export function ProjectDetailsView({
  projectId,
  projectsLabel = 'Projects',
  onNavigateToProjects,
  onDeleteSuccess
}: ProjectDetailsViewProps) {
  const { activeOrganization: organization } = useFrontier();
  // Created per mount: a module-level handle keeps a dialog open across
  // remounts (https://github.com/mui/base-ui/issues/4951).
  // TODO @rohanchkrabrty: This is fixed in Base UI v1.7.0. We can update this after Apsara updates
  const [projectActionsMenuHandle] = useState(() => Menu.createHandle());
  const [deleteProjectDialogHandle] = useState(() =>
    AlertDialog.createHandle<DeleteProjectPayload>()
  );
  const [deleteBlockedDialogHandle] = useState(() =>
    AlertDialog.createHandle<DeleteBlockedPayload>()
  );

  const {
    data: project,
    isLoading: isProjectLoading,
    error: projectError,
    refetch: refetchProject
  } = useQuery(
    FrontierServiceQueries.getProject,
    create(GetProjectRequestSchema, { id: projectId || '' }),
    {
      enabled: !!organization?.id && !!projectId,
      select: d => d?.project
    }
  );

  useEffect(() => {
    if (projectError) {
      toastManager.add({
        title: 'Something went wrong',
        description: projectError.message,
        type: 'error'
      });
    }
  }, [projectError]);

  const {
    data: projectUsersData,
    isLoading: isMembersLoading,
    refetch: refetchProjectUsers
  } = useQuery(
    FrontierServiceQueries.listProjectUsers,
    create(ListProjectUsersRequestSchema, {
      id: projectId || ''
    }),
    { enabled: !!organization?.id && !!projectId }
  );

  const projectUsers = useMemo(
    () => ({
      users: projectUsersData?.users ?? [],
      memberRoles: (projectUsersData?.rolePairs ?? []).reduce(
        (acc: Record<string, ProtoRole[]>, mr: ProjectUserRolePair) => {
          if (mr.userId) acc[mr.userId] = mr.roles;
          return acc;
        },
        {}
      )
    }),
    [projectUsersData]
  );

  const {
    data: projectGroupsData,
    isLoading: isTeamsLoading,
    error: groupsError,
    refetch: refetchProjectGroups
  } = useQuery(
    FrontierServiceQueries.listProjectGroups,
    create(ListProjectGroupsRequestSchema, {
      id: projectId || ''
    }),
    { enabled: !!organization?.id && !!projectId }
  );

  useEffect(() => {
    if (groupsError) {
      toastManager.add({
        title: 'Something went wrong',
        description: groupsError.message,
        type: 'error'
      });
    }
  }, [groupsError]);

  const projectGroups = useMemo(
    () => ({
      groups: projectGroupsData?.groups ?? [],
      groupRoles: (projectGroupsData?.rolePairs ?? []).reduce(
        (acc: Record<string, ProtoRole[]>, gr: ProjectGroupRolePair) => {
          if (gr.groupId) acc[gr.groupId] = gr.roles;
          return acc;
        },
        {}
      )
    }),
    [projectGroupsData]
  );

  const {
    data: rolesData,
    isLoading: isRolesLoading,
    error: rolesError
  } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, {
      state: 'enabled',
      scopes: [PERMISSIONS.ProjectNamespace]
    }),
    { enabled: !!organization?.id && !!projectId }
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

  const resource = `app/project:${projectId}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      { permission: PERMISSIONS.UpdatePermission, resource },
      { permission: PERMISSIONS.DeletePermission, resource }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!projectId
  );

  const { canUpdateProject, canDeleteProject } = useMemo(() => {
    return {
      canUpdateProject: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      ),
      canDeleteProject: shouldShowComponent(
        permissions,
        `${PERMISSIONS.DeletePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const isLoading =
    !organization?.id ||
    isProjectLoading ||
    isMembersLoading ||
    isTeamsLoading ||
    isRolesLoading ||
    isPermissionsFetching;

  const refetchMembers = useCallback(() => {
    refetchProjectUsers();
    refetchProjectGroups();
  }, [refetchProjectUsers, refetchProjectGroups]);

  const [roleFilter, setRoleFilter] = useState('all');

  const members: MemberRow[] = useMemo(() => {
    const teams = projectGroups.groups.map(t => ({
      ...t,
      isTeam: true as const
    }));
    return [...teams, ...projectUsers.users];
  }, [projectGroups.groups, projectUsers.users]);

  const filteredMembers = useMemo(() => {
    if (roleFilter === 'all') return members;
    return members.filter(member => {
      const memberRoleList = member.isTeam
        ? (member.id && projectGroups.groupRoles[member.id]) || []
        : (member.id && projectUsers.memberRoles[member.id]) || [];
      return memberRoleList.some(r => r.id === roleFilter);
    });
  }, [members, roleFilter, projectUsers.memberRoles, projectGroups.groupRoles]);

  const columns = useMemo(
    () =>
      getColumns({
        memberRoles: projectUsers.memberRoles,
        groupRoles: projectGroups.groupRoles,
        roles,
        canUpdateProject,
        menuHandle: memberMenuHandle
      }),
    [
      projectUsers.memberRoles,
      projectGroups.groupRoles,
      roles,
      canUpdateProject
    ]
  );

  const handleDeleteSuccess = useCallback(() => {
    onDeleteSuccess?.();
  }, [onDeleteSuccess]);

  const { startDelete, isChecking, handleDeleteError } = useProjectDeleteCheck({
    menuHandle: projectActionsMenuHandle,
    confirmDialogHandle: deleteProjectDialogHandle,
    blockedDialogHandle: deleteBlockedDialogHandle
  });

  const projectTitle = project?.title || '';
  const projectName = project?.name || '';

  return (
    <ViewContainer>
      <ViewHeader
        title={isProjectLoading ? '' : projectTitle}
        breadcrumb={
          <Breadcrumb size="small">
            <Breadcrumb.Item
              onClick={() => onNavigateToProjects?.()}
            >
              {projectsLabel}
            </Breadcrumb.Item>
            <Breadcrumb.Separator />
            <Breadcrumb.Item current>
              {isProjectLoading ? (
                <Skeleton height="16px" width="100px" />
              ) : (
                projectTitle
              )}
            </Breadcrumb.Item>
          </Breadcrumb>
        }
      >
        {!isLoading && (canUpdateProject || canDeleteProject) && (
          <ProjectActionsMenu
            menuHandle={projectActionsMenuHandle}
            projectId={projectId}
            projectName={projectName}
            projectTitle={projectTitle}
            canUpdate={canUpdateProject}
            canDelete={canDeleteProject}
            isCheckingDelete={isChecking}
            onDelete={() =>
              startDelete({ projectId, projectName: projectTitle })
            }
          />
        )}
      </ViewHeader>

      <DataTable
        data={isLoading ? [] : filteredMembers}
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
                <>
                  <DataTable.Search
                    placeholder="Search by name or email"
                    size="large"
                    width={360}
                  />
                  <Select value={roleFilter} onValueChange={setRoleFilter}>
                    <Select.Trigger className={styles.roleFilter}>
                      <Select.Value placeholder="All" />
                    </Select.Trigger>
                    <Select.Content>
                      <Select.Item value="all">All</Select.Item>
                      {roles.map(role => (
                        <Select.Item key={role.id} value={role.id || ''}>
                          {role.title || role.name}
                        </Select.Item>
                      ))}
                    </Select.Content>
                  </Select>
                </>
              )}
            </Flex>
            {isLoading ? (
              <Skeleton height="34px" width="120px" />
            ) : (
              <AddMemberMenu
                projectId={projectId}
                canUpdateProject={canUpdateProject}
                members={projectUsers.users}
                refetch={refetchMembers}
              />
            )}
          </Flex>
          <DataTable.Content
            emptyState={
              <EmptyState
                icon={<ExclamationTriangleIcon />}
                heading="No members found"
                subHeading="Get started by adding your first member."
              />
            }
            classNames={{
              root: styles.tableRoot,
              table: styles.table,
              row: styles.tableRow
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
                    updateRoleDialogHandle.openWithPayload({
                      memberId: payload.memberId,
                      isTeam: payload.isTeam,
                      role
                    })
                  }
                  data-test-id="frontier-sdk-update-member-role-btn"
                >
                  Make {role.title}
                </Menu.Item>
              ))}
              <Menu.Item
                leadingIcon={<Image src={deleteIcon as unknown as string} alt="Remove" width={16} height={16} />}
                onClick={() =>
                  payload &&
                  removeMemberDialogHandle.openWithPayload({
                    memberId: payload.memberId,
                    projectId,
                    memberType: payload.isTeam ? 'group' : 'user'
                  })
                }
                data-test-id="frontier-sdk-remove-member-btn"
                style={{ color: 'var(--rs-color-foreground-danger-primary)' }}
              >
                Remove
              </Menu.Item>
            </Menu.Content>
          );
        }}
      </Menu>

      <EditProjectDialog
        handle={editProjectDialogHandle}
        refetch={() => {
          refetchProject();
        }}
      />
      <DeleteProjectDialog
        handle={deleteProjectDialogHandle}
        refetch={handleDeleteSuccess}
        onDeleteError={handleDeleteError}
      />
      <DeleteBlockedDialogHost handle={deleteBlockedDialogHandle} />
      <RemoveMemberDialog
        handle={removeMemberDialogHandle}
        refetch={refetchMembers}
      />
      <UpdateRoleDialog
        handle={updateRoleDialogHandle}
        projectId={projectId}
        refetch={refetchMembers}
      />
    </ViewContainer>
  );
}

interface ProjectActionsMenuProps {
  menuHandle: ReturnType<typeof Menu.createHandle>;
  projectId: string;
  projectName: string;
  projectTitle: string;
  canUpdate: boolean;
  canDelete: boolean;
  isCheckingDelete: boolean;
  onDelete: () => void;
}

function ProjectActionsMenu({
  menuHandle,
  projectId,
  projectName,
  projectTitle,
  canUpdate,
  canDelete,
  isCheckingDelete,
  onDelete
}: ProjectActionsMenuProps) {
  return (
    <>
      <Menu.Trigger
        handle={menuHandle}
        render={
          <IconButton
            size={2}
            aria-label="Project actions"
            data-test-id="frontier-sdk-project-details-actions-btn"
          />
        }
      >
        <DotsHorizontalIcon />
      </Menu.Trigger>
      <Menu handle={menuHandle} modal={false}>
        <Menu.Content align="start" className={styles.menuContent}>
          {canUpdate && (
            <Menu.Item
              leadingIcon={<Pencil1Icon />}
              onClick={() =>
                editProjectDialogHandle.openWithPayload({
                  projectId,
                  name: projectName,
                  title: projectTitle
                })
              }
              data-test-id="frontier-sdk-edit-project-details-btn"
            >
              Edit
            </Menu.Item>
          )}
          {canDelete && (
            <DeleteProjectMenuItem
              isChecking={isCheckingDelete}
              onClick={onDelete}
              data-test-id="frontier-sdk-delete-project-details-btn"
            />
          )}
        </Menu.Content>
      </Menu>
    </>
  );
}
