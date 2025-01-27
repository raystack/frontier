import { DataTable, Dialog, Image, Separator } from '@raystack/apsara';
import styles from './styles.module.css';
import { Checkbox, Flex, Text } from '@raystack/apsara/v1';
import { useNavigate, useParams } from '@tanstack/react-router';
import cross from '~/react/assets/cross.svg';
import { useEffect, useState } from 'react';
import { V1Beta1Project } from '~/src';
import { useFrontier } from '~/react/contexts/FrontierContext';

function getColumns() {
  return [
    {
      header: '',
      accessorKey: 'id',
      enableSorting: false,
      cell: ({}) => {
        return (
          <Flex>
            <Checkbox />
          </Flex>
        );
      }
    },
    {
      header: 'Project Name',
      accessorKey: 'title',
      cell: ({ getValue }) => {
        const value = getValue();
        return <Flex direction="column">{value}</Flex>;
      }
    },
    {
      header: 'Access',
      accessorKey: 'id',
      enableSorting: false,
      meta: {
        style: {
          padding: 0
        }
      },
      cell: <Flex></Flex>
    }
  ];
}

export default function ManageServiceUserProjects() {
  const { client, activeOrganization: organization } = useFrontier();

  const { id } = useParams({
    from: '/api-keys/$id/projects'
  });

  const [projects, setProjects] = useState<V1Beta1Project[]>([]);
  const [isProjectsLoading, setIsProjectsLoading] = useState(false);
  const [isAddedProjectsLoading, setIsAddedProjectsLoading] = useState(false);

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
        const data = await client?.frontierServiceListProjectsByUser(id);
        console.log(data);
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
          (a?.title || '') > (b?.title || '') ? 1 : -1
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

  const columns = getColumns();

  const data = projects || [];

  const isLoading = isProjectsLoading || isAddedProjectsLoading;

  return (
    <Dialog open={true}>
      <Dialog.Content
        overlayClassname={styles.overlay}
        className={styles.manageProjectDialogContent}
      >
        <Flex justify="between" className={styles.manageProjectDialog}>
          <Text size={6} weight={500}>
            Manage Project Access
          </Text>

          <Image
            alt="cross"
            style={{ cursor: 'pointer' }}
            // @ts-ignore
            src={cross}
            onClick={onCancel}
            data-test-id="frontier-sdk-service-account-manage-access-close-btn"
          />
        </Flex>
        <Separator />
        <Flex
          className={styles.manageProjectDialogWrapper}
          gap="large"
          direction={'column'}
        >
          <Text size={2} variant={'secondary'}>
            Note: Choose a project to join and specify the access type.
          </Text>
          <DataTable columns={columns} data={data} isLoading={isLoading} />
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}
