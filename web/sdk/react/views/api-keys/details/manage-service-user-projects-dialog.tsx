import {
  Checkbox,
  Flex,
  Spinner,
  Text,
  toast,
  Image,
  Dialog,
  DataTable,
  type DataTableColumnDef
} from '@raystack/apsara';
import { useCallback, useState, useEffect, useMemo } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { PERMISSIONS } from '~/utils';
import cross from '~/react/assets/cross.svg';
import styles from './service-user.module.css';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  ListServiceUserProjectsRequestSchema,
  ListOrganizationProjectsRequestSchema,
  ListRolesRequestSchema,
  SetProjectMemberRoleRequestSchema,
  RemoveProjectMemberRequestSchema,
  Project
} from '@raystack/proton/frontier';
import { orderBy } from 'lodash';

type ProjectAccessMap = Record<string, { value: boolean; isLoading: boolean }>;

const getColumns = ({
  permMap,
  onChange
}: {
  permMap: ProjectAccessMap;
  onChange: (projectId: string, value: boolean) => void;
}): DataTableColumnDef<Project, unknown>[] => {
  return [
    {
      header: '',
      id: 'checkbox',
      accessorKey: 'id',
      enableSorting: false,
      styles: {
        cell: {
          width: 'var(--rs-space-2)'
        }
      },
      cell: ({ getValue }) => {
        const projectId = getValue() as string;
        const { value, isLoading } = permMap[projectId] || {};
        return (
          <Flex>
            {isLoading ? (
              <Spinner />
            ) : (
              <Checkbox
                checked={value}
                onCheckedChange={v => onChange(projectId, v === true)}
              />
            )}
          </Flex>
        );
      }
    },
    {
      header: 'Project Name',
      accessorKey: 'title',
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return (
          <Flex direction="column">
            <Text>{value}</Text>
          </Flex>
        );
      }
    },
    {
      header: 'Access',
      accessorKey: 'id',
      enableSorting: false,
      cell: () => (
        <Flex>
          <Text>Member</Text>
        </Flex>
      )
    }
  ];
};

export interface ManageServiceUserProjectsDialogProps {
  open: boolean;
  onOpenChange?: (value: boolean) => void;
  serviceUserId: string;
}

export default function ManageServiceUserProjectsDialog({
  open,
  onOpenChange,
  serviceUserId
}: ManageServiceUserProjectsDialogProps) {
  const { activeOrganization: organization } = useFrontier();

  const [addedProjectsMap, setAddedProjectsMap] = useState<ProjectAccessMap>(
    {}
  );

  const handleClose = () => onOpenChange?.(false);

  const orgId = organization?.id || '';

  const { data: projectsData, isLoading: isProjectsLoading } = useQuery(
    FrontierServiceQueries.listOrganizationProjects,
    create(ListOrganizationProjectsRequestSchema, {
      id: orgId,
      state: '',
      withMemberCount: false
    }),
    {
      enabled: Boolean(orgId) && open
    }
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
      {
        enabled: Boolean(serviceUserId) && Boolean(orgId) && open
      }
    );

  const addedProjects = useMemo(
    () => addedProjectsData?.projects ?? [],
    [addedProjectsData]
  );

  // Initialize addedProjectsMap from query data
  useEffect(() => {
    const permMap = addedProjects.reduce((acc, proj) => {
      acc[proj?.id || ''] = { value: true, isLoading: false };
      return acc;
    }, {} as ProjectAccessMap);
    setAddedProjectsMap(permMap);
  }, [addedProjects]);

  const { data: rolesData } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, {
      state: 'enabled',
      scopes: [PERMISSIONS.ProjectNamespace]
    }),
    { enabled: open }
  );

  const ownerRoleId = useMemo(
    () => rolesData?.roles?.find(r => r.name === PERMISSIONS.RoleProjectOwner)?.id ?? '',
    [rolesData]
  );

  const { mutateAsync: setProjectMemberRole } = useMutation(
    FrontierServiceQueries.setProjectMemberRole
  );

  const { mutateAsync: removeProjectMember } = useMutation(
    FrontierServiceQueries.removeProjectMember
  );

  const onAccessChange = useCallback(
    async (projectId: string, value: boolean) => {
      try {
        setAddedProjectsMap(prev => ({
          ...prev,
          [projectId]: { ...prev[projectId], isLoading: true }
        }));

        if (value) {
          if (!ownerRoleId) throw new Error('Project owner role not found');
          await setProjectMemberRole(
            create(SetProjectMemberRoleRequestSchema, {
              projectId,
              principalId: serviceUserId,
              principalType: PERMISSIONS.ServiceUserPrincipal,
              roleId: ownerRoleId
            })
          );
          setAddedProjectsMap(prev => ({
            ...prev,
            [projectId]: { value: true, isLoading: false }
          }));
        } else {
          await removeProjectMember(
            create(RemoveProjectMemberRequestSchema, {
              projectId,
              principalId: serviceUserId,
              principalType: PERMISSIONS.ServiceUserPrincipal
            })
          );
          setAddedProjectsMap(prev => ({
            ...prev,
            [projectId]: { value: false, isLoading: false }
          }));
        }
      } catch (error: unknown) {
        console.error(error);
        toast.error('unable to update project access');
        setAddedProjectsMap(prev => ({
          ...prev,
          [projectId]: { ...prev[projectId], isLoading: false }
        }));
      }
    },
    [serviceUserId, setProjectMemberRole, removeProjectMember]
  );

  const columns = getColumns({
    permMap: addedProjectsMap,
    onChange: onAccessChange
  });

  const data = projects || [];

  const isLoading = isProjectsLoading || isAddedProjectsLoading;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Content
        overlayClassName={styles.overlay}
        className={styles.manageProjectDialogContent}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Manage Project Access
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={handleClose}
              data-test-id="frontier-sdk-service-account-manage-access-close-btn"
            />
          </Flex>
        </Dialog.Header>

        <Dialog.Body>
          <Flex
            className={styles.manageProjectDialogWrapper}
            gap={9}
            direction={'column'}
          >
            <Text size="small" variant="secondary">
              Note: Select projects to give access to the service user.
            </Text>
            <DataTable
              columns={columns}
              data={data}
              isLoading={isLoading}
              mode="client"
              defaultSort={{ name: 'name', order: 'asc' }}
            >
              <DataTable.Content classNames={{ root: styles.tableRoot }} />
            </DataTable>
          </Flex>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
}
