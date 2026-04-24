'use client';

import { useCallback, useState, useEffect, useMemo } from 'react';
import {
  Checkbox,
  Flex,
  Spinner,
  Text,
  Dialog,
  DataTable,
  Select,
  toastManager
} from '@raystack/apsara-v1';
import type { DataTableColumnDef } from '@raystack/apsara-v1';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  ListServiceUserProjectsRequestSchema,
  ListOrganizationProjectsRequestSchema,
  ListRolesRequestSchema,
  SetProjectMemberRoleRequestSchema,
  RemoveProjectMemberRequestSchema,
  type Project
} from '@raystack/proton/frontier';
import { orderBy } from 'lodash';
import { useFrontier } from '../../../contexts/FrontierContext';
import { PERMISSIONS } from '../../../../utils';
import styles from '../service-account-details-view.module.css';

const PROJECT_ROLES = [
  { value: PERMISSIONS.RoleProjectViewer, label: 'Viewer' },
  { value: PERMISSIONS.RoleProjectOwner, label: 'Owner' }
];

type ProjectAccessMap = Record<
  string,
  { value: boolean; isLoading: boolean; roleId: string }
>;

function getColumns({
  permMap,
  onCheckChange,
  onRoleChange
}: {
  permMap: ProjectAccessMap;
  onCheckChange: (projectId: string, value: boolean) => void;
  onRoleChange: (projectId: string, roleId: string) => void;
}): DataTableColumnDef<Project, unknown>[] {
  return [
    {
      header: '',
      id: 'checkbox',
      accessorKey: 'id',
      enableSorting: false,
      cell: ({ getValue }) => {
        const projectId = getValue() as string;
        const entry = permMap[projectId];
        const isLoading = entry?.isLoading;
        const checked = entry?.value ?? false;
        return (
          <Flex>
            {isLoading ? (
              <Spinner />
            ) : (
              <Checkbox
                checked={checked}
                onCheckedChange={v => onCheckChange(projectId, v === true)}
              />
            )}
          </Flex>
        );
      }
    },
    {
      header: 'Project Name',
      accessorKey: 'title',
      cell: ({ getValue }) => (
        <Text size="regular">{getValue() as string}</Text>
      )
    },
    {
      header: 'Access',
      id: 'access',
      accessorKey: 'id',
      enableSorting: false,
      cell: ({ getValue }) => {
        const projectId = getValue() as string;
        const entry = permMap[projectId];
        const isChecked = entry?.value ?? false;
        const roleId = entry?.roleId || PERMISSIONS.RoleProjectViewer;
        return (
          <Select
            value={roleId}
            onValueChange={val => onRoleChange(projectId, val)}
            disabled={!isChecked}
          >
            <Select.Trigger className={styles.accessSelectTrigger}>
              <Select.Value placeholder="Viewer" />
            </Select.Trigger>
            <Select.Content>
              {PROJECT_ROLES.map(role => (
                <Select.Item key={role.value} value={role.value}>
                  {role.label}
                </Select.Item>
              ))}
            </Select.Content>
          </Select>
        );
      }
    }
  ];
}

export interface ManageProjectAccessDialogProps {
  handle: ReturnType<typeof Dialog.createHandle>;
  serviceUserId: string;
}

export function ManageProjectAccessDialog({
  handle,
  serviceUserId
}: ManageProjectAccessDialogProps) {
  const { activeOrganization: organization } = useFrontier();
  const orgId = organization?.id || '';

  const [isOpen, setIsOpen] = useState(false);
  const [accessMap, setAccessMap] = useState<ProjectAccessMap>({});

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open);
  };

  const { data: projectsData, isLoading: isProjectsLoading } = useQuery(
    FrontierServiceQueries.listOrganizationProjects,
    create(ListOrganizationProjectsRequestSchema, {
      id: orgId,
      state: '',
      withMemberCount: false
    }),
    { enabled: Boolean(orgId) && isOpen }
  );

  const projects = useMemo(() => {
    const list = projectsData?.projects ?? [];
    return orderBy(list, ['title'], ['asc']);
  }, [projectsData]);

  const { data: addedProjectsData, isLoading: isAddedProjectsLoading } =
    useQuery(
      FrontierServiceQueries.listServiceUserProjects,
      create(ListServiceUserProjectsRequestSchema, {
        id: serviceUserId,
        orgId,
        withPermissions: []
      }),
      { enabled: Boolean(serviceUserId) && Boolean(orgId) && isOpen }
    );

  const addedProjects = useMemo(
    () => addedProjectsData?.projects ?? [],
    [addedProjectsData]
  );

  const { data: rolesData } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, {
      state: 'enabled',
      scopes: [PERMISSIONS.ProjectNamespace]
    }),
    { enabled: Boolean(orgId) && isOpen }
  );

  const roleNameToId = useMemo(() => {
    const map: Record<string, string> = {};
    for (const r of rolesData?.roles ?? []) {
      if (r.name && r.id) map[r.name] = r.id;
    }
    return map;
  }, [rolesData]);

  useEffect(() => {
    const permMap = addedProjects.reduce((acc, proj) => {
      acc[proj?.id || ''] = {
        value: true,
        isLoading: false,
        roleId: PERMISSIONS.RoleProjectOwner
      };
      return acc;
    }, {} as ProjectAccessMap);
    setAccessMap(permMap);
  }, [addedProjects]);

  const { mutateAsync: setProjectMemberRole } = useMutation(
    FrontierServiceQueries.setProjectMemberRole
  );

  const { mutateAsync: removeProjectMember } = useMutation(
    FrontierServiceQueries.removeProjectMember
  );

  const onCheckChange = useCallback(
    async (projectId: string, value: boolean) => {
      try {
        setAccessMap(prev => ({
          ...prev,
          [projectId]: {
            ...prev[projectId],
            isLoading: true,
            roleId: prev[projectId]?.roleId || PERMISSIONS.RoleProjectViewer
          }
        }));

        if (value) {
          const roleName =
            accessMap[projectId]?.roleId || PERMISSIONS.RoleProjectViewer;
          const resolvedRoleId = roleNameToId[roleName] || '';
          await setProjectMemberRole(
            create(SetProjectMemberRoleRequestSchema, {
              projectId,
              principalId: serviceUserId,
              principalType: PERMISSIONS.ServiceUserPrincipal,
              roleId: resolvedRoleId
            })
          );
          setAccessMap(prev => ({
            ...prev,
            [projectId]: { value: true, isLoading: false, roleId: roleName }
          }));
        } else {
          await removeProjectMember(
            create(RemoveProjectMemberRequestSchema, {
              projectId,
              principalId: serviceUserId,
              principalType: PERMISSIONS.ServiceUserPrincipal
            })
          );
          setAccessMap(prev => ({
            ...prev,
            [projectId]: {
              value: false,
              isLoading: false,
              roleId: prev[projectId]?.roleId || PERMISSIONS.RoleProjectViewer
            }
          }));
        }
      } catch (error: unknown) {
        toastManager.add({
          title: 'Unable to update project access',
          description: error instanceof Error ? error.message : 'Unknown error',
          type: 'error'
        });
        setAccessMap(prev => ({
          ...prev,
          [projectId]: { ...prev[projectId], isLoading: false }
        }));
      }
    },
    [
      serviceUserId,
      accessMap,
      roleNameToId,
      setProjectMemberRole,
      removeProjectMember
    ]
  );

  const onRoleChange = useCallback(
    async (projectId: string, roleId: string) => {
      const entry = accessMap[projectId];
      if (!entry?.value) {
        setAccessMap(prev => ({
          ...prev,
          [projectId]: { ...prev[projectId], roleId }
        }));
        return;
      }

      try {
        setAccessMap(prev => ({
          ...prev,
          [projectId]: { ...prev[projectId], isLoading: true, roleId }
        }));

        const resolvedRoleId = roleNameToId[roleId] || '';
        await setProjectMemberRole(
          create(SetProjectMemberRoleRequestSchema, {
            projectId,
            principalId: serviceUserId,
            principalType: PERMISSIONS.ServiceUserPrincipal,
            roleId: resolvedRoleId
          })
        );

        setAccessMap(prev => ({
          ...prev,
          [projectId]: { value: true, isLoading: false, roleId }
        }));
      } catch (error: unknown) {
        toastManager.add({
          title: 'Unable to update project role',
          description: error instanceof Error ? error.message : 'Unknown error',
          type: 'error'
        });
        setAccessMap(prev => ({
          ...prev,
          [projectId]: { ...prev[projectId], isLoading: false }
        }));
      }
    },
    [
      serviceUserId,
      accessMap,
      roleNameToId,
      setProjectMemberRole
    ]
  );

  const columns = useMemo(
    () =>
      getColumns({
        permMap: accessMap,
        onCheckChange,
        onRoleChange
      }),
    [accessMap, onCheckChange, onRoleChange]
  );

  const isLoading = isProjectsLoading || isAddedProjectsLoading;

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      <Dialog.Content width={1024} className={styles.manageAccessDialogContent}>
        <Dialog.Header>
          <Dialog.Title>Manage Project Access</Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className={styles.manageAccessDialogBody}>
          <Flex direction="column" gap={7}>
            <Text size="small" variant="secondary">
              Note: Choose a project to join and specify the access type.
            </Text>
            <DataTable
              data={projects}
              columns={columns}
              isLoading={isLoading}
              defaultSort={{ name: 'title', order: 'asc' }}
              mode="client"
            >
              <DataTable.Content
                classNames={{ root: styles.manageAccessTableRoot }}
              />
            </DataTable>
          </Flex>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
}
