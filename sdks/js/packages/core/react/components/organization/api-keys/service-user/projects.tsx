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
import { useNavigate, useParams } from '@tanstack/react-router';
import { useCallback, useState, useEffect } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { PERMISSIONS } from '~/utils';
import cross from '~/react/assets/cross.svg';
import styles from './styles.module.css';
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
  Project,
  Policy
} from '@raystack/proton/frontier';

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

export default function ManageServiceUserProjects() {
  const { activeOrganization: organization } = useFrontier();

  const { id } = useParams({
    from: '/api-keys/$id/projects'
  });

  const [addedProjectsMap, setAddedProjectsMap] = useState<ProjectAccessMap>(
    {}
  );

  const navigate = useNavigate({ from: '/api-keys/$id/projects' });

  const orgId = organization?.id || '';

  const {
    data: projects = [],
    isLoading: isProjectsLoading
  } = useQuery(
    FrontierServiceQueries.listOrganizationProjects,
    create(ListOrganizationProjectsRequestSchema, {
      id: orgId,
      state: '',
      withMemberCount: false
    }),
    {
      enabled: Boolean(orgId),
      select: data => {
        const list = data?.projects ?? [];
        return list.sort((a, b) =>
          (a?.title?.toLowerCase() || '') > (b?.title?.toLowerCase() || '')
            ? 1
            : -1
        );
      }
    }
  );

  const {
    data: addedProjects = [],
    isLoading: isAddedProjectsLoading
  } = useQuery(
    FrontierServiceQueries.listServiceUserProjects,
    create(ListServiceUserProjectsRequestSchema, {
      id,
      orgId,
      withPermissions: []
    }),
    {
      enabled: Boolean(id) && Boolean(orgId),
      select: data => data?.projects ?? []
    }
  );

  // Initialize addedProjectsMap from query data
  useEffect(() => {
    const permMap = addedProjects.reduce((acc, proj) => {
      acc[proj?.id || ''] = { value: true, isLoading: false };
      return acc;
    }, {} as ProjectAccessMap);
    setAddedProjectsMap(permMap);
  }, [addedProjects]);

  function onCancel() {
    navigate({
      to: '/api-keys/$id',
      params: {
        id: id
      }
    });
  }

  const { mutateAsync: createPolicyForProject } = useMutation(
    FrontierServiceQueries.createPolicyForProject
  );

  const { mutateAsync: listPolicies } = useMutation(
    FrontierServiceQueries.listPolicies
  );

  const { mutateAsync: deletePolicy } = useMutation(
    FrontierServiceQueries.deletePolicy
  );

  const onAccessChange = useCallback(
    async (projectId: string, value: boolean) => {
      try {
        setAddedProjectsMap(prev => ({
          ...prev,
          [projectId]: { ...prev[projectId], isLoading: true }
        }));

        if (value) {
          const principal = `${PERMISSIONS.ServiceUserPrincipal}:${id}`;
          await createPolicyForProject(
            create(CreatePolicyForProjectRequestSchema, {
              projectId,
              body: create(CreatePolicyForProjectBodySchema, {
                roleId: PERMISSIONS.RoleProjectOwner,
                principal
              })
            })
          );
          setAddedProjectsMap(prev => ({
            ...prev,
            [projectId]: { value: true, isLoading: false }
          }));
        } else {
          const policiesResp = await listPolicies(
            create(ListPoliciesRequestSchema, {
              projectId,
              userId: id,
              orgId: '',
              roleId: '',
              groupId: ''
            })
          );
          const policies = policiesResp?.policies || [];
          const deletePromises = policies.map((p: Policy) =>
            deletePolicy(
              create(DeletePolicyRequestSchema, {
                id: p.id
              })
            )
          );
          await Promise.all(deletePromises);
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
    [id, createPolicyForProject, listPolicies, deletePolicy]
  );

  const columns = getColumns({
    permMap: addedProjectsMap,
    onChange: onAccessChange
  });

  const data = projects || [];

  const isLoading = isProjectsLoading || isAddedProjectsLoading;

  return (
    <Dialog open={true}>
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
              onClick={onCancel}
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
