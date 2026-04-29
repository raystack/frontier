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
  CreatePolicyForProjectRequestSchema,
  CreatePolicyForProjectBodySchema,
  ListPoliciesRequestSchema,
  DeletePolicyRequestSchema,
  type Project,
  type Policy
} from '@raystack/proton/frontier';
import { orderBy } from 'lodash';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { PERMISSIONS } from '~/utils';
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
      styles: {
        cell: { width: '40px', minWidth: '40px', maxWidth: '40px' },
        header: { width: '40px', minWidth: '40px', maxWidth: '40px' }
      },
      cell: ({ getValue }) => {
        const projectId = getValue() as string;
        const entry = permMap[projectId];
        const isLoading = entry?.isLoading;
        const checked = entry?.value ?? false;
        return (
          <Flex>
            {isLoading ? (
              <Spinner size={2} />
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
      styles: {
        cell: { width: '180px', minWidth: '180px', maxWidth: '180px' },
        header: { width: '180px', minWidth: '180px', maxWidth: '180px' }
      },
      cell: ({ getValue }) => {
        const projectId = getValue() as string;
        const entry = permMap[projectId];
        const isChecked = entry?.value ?? false;
        const isLoading = entry?.isLoading ?? false;
        const roleId = entry?.roleId || PERMISSIONS.RoleProjectViewer;
        return (
          <Select
            value={roleId}
            onValueChange={val => onRoleChange(projectId, val)}
            disabled={!isChecked || isLoading}
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

  const { mutateAsync: createPolicyForProject } = useMutation(
    FrontierServiceQueries.createPolicyForProject
  );

  const { mutateAsync: listPolicies } = useMutation(
    FrontierServiceQueries.listPolicies
  );

  const { mutateAsync: deletePolicy } = useMutation(
    FrontierServiceQueries.deletePolicy
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
          const roleId =
            accessMap[projectId]?.roleId || PERMISSIONS.RoleProjectViewer;
          const principal = `${PERMISSIONS.ServiceUserPrincipal}:${serviceUserId}`;
          await createPolicyForProject(
            create(CreatePolicyForProjectRequestSchema, {
              projectId,
              body: create(CreatePolicyForProjectBodySchema, {
                roleId,
                principal
              })
            })
          );
          setAccessMap(prev => ({
            ...prev,
            [projectId]: { value: true, isLoading: false, roleId }
          }));
        } else {
          const policiesResp = await listPolicies(
            create(ListPoliciesRequestSchema, {
              projectId,
              userId: serviceUserId,
              orgId: '',
              roleId: '',
              groupId: ''
            })
          );
          const policies = policiesResp?.policies || [];
          await Promise.all(
            policies.map((p: Policy) =>
              deletePolicy(create(DeletePolicyRequestSchema, { id: p.id }))
            )
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
      createPolicyForProject,
      listPolicies,
      deletePolicy
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

        const policiesResp = await listPolicies(
          create(ListPoliciesRequestSchema, {
            projectId,
            userId: serviceUserId,
            orgId: '',
            roleId: '',
            groupId: ''
          })
        );
        const policies = policiesResp?.policies || [];
        await Promise.all(
          policies.map((p: Policy) =>
            deletePolicy(create(DeletePolicyRequestSchema, { id: p.id }))
          )
        );

        const principal = `${PERMISSIONS.ServiceUserPrincipal}:${serviceUserId}`;
        await createPolicyForProject(
          create(CreatePolicyForProjectRequestSchema, {
            projectId,
            body: create(CreatePolicyForProjectBodySchema, {
              roleId,
              principal
            })
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
      createPolicyForProject,
      listPolicies,
      deletePolicy
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
          <Flex
            direction="column"
            gap={7}
            className={styles.manageAccessBodyInner}
          >
            <Text size="small" variant="secondary">
              Note: Choose a project to join and specify the access type.
            </Text>
            <DataTable
              data={projects}
              columns={columns}
              isLoading={isLoading}
              loadingRowCount={9}
              defaultSort={{ name: 'title', order: 'asc' }}
              mode="client"
            >
              <DataTable.Content
                classNames={{
                  root: styles.manageAccessTableRoot
                }}
              />
            </DataTable>
          </Flex>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
}
