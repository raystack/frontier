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
import { useCallback, useEffect, useState } from 'react';
import type {
  V1Beta1CreatePolicyForProjectBody,
  V1Beta1Policy,
  V1Beta1Project
} from '~/src';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { PERMISSIONS } from '~/utils';
import cross from '~/react/assets/cross.svg';
import styles from './styles.module.css';

type ProjectAccessMap = Record<string, { value: boolean; isLoading: boolean }>;

const getColumns = ({
  permMap,
  onChange
}: {
  permMap: ProjectAccessMap;
  onChange: (projectId: string, value: boolean) => void;
}): DataTableColumnDef<V1Beta1Project, unknown>[] => {
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
        const projectId = getValue();
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
        const value = getValue();
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
          <Text>Viewer</Text>
        </Flex>
      )
    }
  ];
};

export default function ManageServiceUserProjects() {
  const { client, activeOrganization: organization } = useFrontier();

  const { id } = useParams({
    from: '/api-keys/$id/projects'
  });

  const [projects, setProjects] = useState<V1Beta1Project[]>([]);
  const [isProjectsLoading, setIsProjectsLoading] = useState(false);
  const [isAddedProjectsLoading, setIsAddedProjectsLoading] = useState(false);
  const [addedProjectsMap, setAddedProjectsMap] = useState<ProjectAccessMap>(
    {}
  );

  const navigate = useNavigate({ from: '/api-keys/$id/projects' });

  const orgId = organization?.id || '';

  function onCancel() {
    navigate({
      to: '/api-keys/$id',
      params: {
        id: id
      }
    });
  }

  useEffect(() => {
    async function fetchAddedProjects() {
      try {
        setIsAddedProjectsLoading(true);
        const data = await client?.frontierServiceListServiceUserProjects(
          orgId,
          id
        );
        const permMap = data?.data?.projects?.reduce((acc, proj) => {
          acc[proj?.id || ''] = { value: true, isLoading: false };
          return acc;
        }, {} as ProjectAccessMap);
        setAddedProjectsMap(permMap || {});
      } catch (error: unknown) {
        console.error(error);
      } finally {
        setIsAddedProjectsLoading(false);
      }
    }

    async function fetchProjects() {
      try {
        setIsProjectsLoading(true);
        const data = await client?.frontierServiceListOrganizationProjects(
          orgId
        );
        const list = data?.data?.projects?.sort((a, b) =>
          (a?.title?.toLowerCase() || '') > (b?.title?.toLowerCase() || '')
            ? 1
            : -1
        );
        setProjects(list || []);
      } catch (error: unknown) {
        console.error(error);
      } finally {
        setIsProjectsLoading(false);
      }
    }
    if (orgId) {
      fetchProjects();
      fetchAddedProjects();
    }
  }, [client, id, orgId]);

  const onAccessChange = useCallback(
    async (projectId: string, value: boolean) => {
      try {
        setAddedProjectsMap(prev => ({
          ...prev,
          [projectId]: { ...prev[projectId], isLoading: true }
        }));

        if (value) {
          const principal = `${PERMISSIONS.ServiceUserPrincipal}:${id}`;
          const policy: V1Beta1CreatePolicyForProjectBody = {
            role_id: PERMISSIONS.RoleProjectViewer,
            principal
          };
          await client?.frontierServiceCreatePolicyForProject(
            projectId,
            policy
          );
          setAddedProjectsMap(prev => ({
            ...prev,
            [projectId]: { value: true, isLoading: false }
          }));
        } else {
          const policiesResp = await client?.frontierServiceListPolicies({
            project_id: projectId,
            user_id: id
          });
          const policies = policiesResp?.data?.policies || [];
          const deletePromises = policies.map((p: V1Beta1Policy) =>
            client?.frontierServiceDeletePolicy(p.id as string)
          );
          await Promise.all(deletePromises);
          setAddedProjectsMap(prev => ({
            ...prev,
            [projectId]: { value: false, isLoading: false }
          }));
        }
      } catch (err: unknown) {
        console.error(err);
        toast.error('unable to update project access');
        setAddedProjectsMap(prev => ({
          ...prev,
          [projectId]: { ...prev[projectId], isLoading: false }
        }));
      }
    },
    [client, id]
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
        overlayClassname={styles.overlay}
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
