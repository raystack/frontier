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
} from '@raystack/apsara-v1';
import deleteIcon from '../../assets/delete.svg';
import { toastManager } from '@raystack/apsara-v1';
import {
  useQuery,
  useMutation
} from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  ListProjectGroupsRequestSchema,
  ListProjectUsersRequestSchema,
  GetProjectRequestSchema,
  ListRolesRequestSchema,
  SetProjectMemberRoleRequestSchema,
  type Role as ProtoRole
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { useFrontier } from '../../contexts/FrontierContext';
import { usePermissions } from '../../hooks/usePermissions';
import { PERMISSIONS, shouldShowComponent } from '../../../utils';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import {
  EditProjectDialog,
  type EditProjectPayload
} from './components/edit-project-dialog';
import {
  DeleteProjectDialog,
  type DeleteProjectPayload
} from './components/delete-project-dialog';
import {
  RemoveMemberDialog,
  type RemoveMemberPayload
} from './components/remove-member-dialog';
import {
  getColumns,
  type MemberRow,
  type MemberMenuPayload
} from './components/member-columns';
import { AddMemberMenu } from './components/add-member-menu';
import styles from './project-details-view.module.css';
import { handleConnectError } from '~/utils/error';

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
const deleteProjectDialogHandle =
  AlertDialog.createHandle<DeleteProjectPayload>();
const removeMemberDialogHandle =
  AlertDialog.createHandle<RemoveMemberPayload>();

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
      id: projectId || '',
      withRoles: true
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
      id: projectId || '',
      withRoles: true
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

  const { mutateAsync: setProjectMemberRole } = useMutation(
    FrontierServiceQueries.setProjectMemberRole
  );

  const updateMemberRole = useCallback(
    async (memberId: string, isTeam: boolean, role: ProtoRole) => {
      try {
        await setProjectMemberRole(
          create(SetProjectMemberRoleRequestSchema, {
            projectId,
            principalId: memberId,
            principalType: isTeam ? PERMISSIONS.GroupNamespace : PERMISSIONS.UserNamespace,
            roleId: role.id as string
          })
        );
        refetchMembers();
        toastManager.add({
          title: 'Member role updated',
          type: 'success'
        });
      } catch (error) {
        handleConnectError(error, {
          PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
          NotFound: (err) => toastManager.add({ title: 'Not found', description: err.message, type: 'error' }),
          Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
        });
      }
    },
    [setProjectMemberRole, projectId, refetchMembers]
  );

  const handleDeleteSuccess = useCallback(() => {
    onDeleteSuccess?.();
  }, [onDeleteSuccess]);

  const projectTitle = project?.title || '';

  return (
    <ViewContainer>
      <ViewHeader
        title={isProjectLoading ? '' : projectTitle}
        breadcrumb={
          <Breadcrumb size="small">
            <Breadcrumb.Item
              href="#"
              onClick={(e) => {
                e.preventDefault();
                onNavigateToProjects?.();
              }}
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
            projectId={projectId}
            projectTitle={projectTitle}
            canUpdate={canUpdateProject}
            canDelete={canDeleteProject}
          />
        )}
      </ViewHeader>

      <DataTable
        data={filteredMembers}
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
                    updateMemberRole(
                      payload.memberId,
                      payload.isTeam,
                      role
                    )
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
                    projectId
                  })
                }
                data-test-id="frontier-sdk-remove-member-btn"
                style={{ color: 'var(--rs-color-foreground-danger-primary)' }}
              >
                Remove from project
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
      />
      <RemoveMemberDialog
        handle={removeMemberDialogHandle}
        refetch={refetchMembers}
      />
    </ViewContainer>
  );
}

interface ProjectActionsMenuProps {
  projectId: string;
  projectTitle: string;
  canUpdate: boolean;
  canDelete: boolean;
}

const projectActionsMenuHandle = Menu.createHandle();

function ProjectActionsMenu({
  projectId,
  projectTitle,
  canUpdate,
  canDelete
}: ProjectActionsMenuProps) {
  return (
    <>
      <Menu.Trigger
        handle={projectActionsMenuHandle}
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
      <Menu handle={projectActionsMenuHandle} modal={false}>
        <Menu.Content align="start" className={styles.menuContent}>
          {canUpdate && (
            <Menu.Item
              leadingIcon={<Pencil1Icon />}
              onClick={() =>
                editProjectDialogHandle.openWithPayload({
                  projectId,
                  title: projectTitle
                })
              }
              data-test-id="frontier-sdk-edit-project-details-btn"
            >
              Edit
            </Menu.Item>
          )}
          {canDelete && (
            <Menu.Item
              leadingIcon={<Image src={deleteIcon as unknown as string} alt="Delete" width={16} height={16} />}
              onClick={() =>
                deleteProjectDialogHandle.openWithPayload({ projectId })
              }
              data-test-id="frontier-sdk-delete-project-details-btn"
              style={{ color: 'var(--rs-color-foreground-danger-primary)' }}
            >
              Delete project
            </Menu.Item>
          )}
        </Menu.Content>
      </Menu>
    </>
  );
}
