import {
  Button,
  Checkbox,
  toast,
  Skeleton,
  Image,
  Text,
  Flex,
  Dialog,
  InputField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries, GetProjectRequestSchema, DeleteProjectRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import cross from '~/react/assets/cross.svg';
import styles from '../organization.module.css';

const projectSchema = yup
  .object({
    name: yup.string()
  })
  .required();

export const DeleteProject = () => {
  const {
    watch,
    setError,
    handleSubmit,
    formState: { errors, isSubmitting },
    register
  } = useForm({
    resolver: yupResolver(projectSchema)
  });
  let { projectId } = useParams({ from: '/projects/$projectId/delete' });
  const navigate = useNavigate({ from: '/projects/$projectId/delete' });
  const { activeOrganization: organization } = useFrontier();
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  const {
    data: project,
    isLoading: isProjectQueryLoading,
    error: projectError
  } = useQuery(
    FrontierServiceQueries.getProject,
    create(GetProjectRequestSchema, { id: projectId || '' }),
    {
      enabled: !!projectId,
      select: (d) => d?.project
    }
  );

  useEffect(() => {
    if (projectError) {
      toast.error('Something went wrong', { description: projectError.message });
    }
  }, [projectError]);

  const { mutateAsync: deleteProject } = useMutation(
    FrontierServiceQueries.deleteProject,
    {
      onSuccess: () => {
        toast.success('project deleted');
        navigate({ to: '/projects' });
      },
      onError: (err: Error) =>
        toast.error('Something went wrong', { description: err.message })
    }
  );

  async function onSubmit(data: { name?: string }) {
    if (!organization?.id || !projectId) return;
    if (data.name !== project?.name)
      return setError('name', { message: 'project name is not same' });
    await deleteProject(create(DeleteProjectRequestSchema, { id: projectId }));
  }

  const name = watch('name', '');
  return (
    <Dialog open={true}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
        overlayClassName={styles.overlay}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Verify project deletion
            </Text>
            <Image
              alt="cross"
              src={cross as unknown as string}
              onClick={() =>
                navigate({ to: '/projects/$projectId', params: { projectId } })
              }
              style={{ cursor: 'pointer' }}
              data-test-id="frontier-sdk-delete-project-close-btn"
            />
          </Flex>
        </Dialog.Header>
        <Dialog.Body>
          <form onSubmit={handleSubmit(onSubmit)}>
            <Flex direction="column" gap={5}>
              {isProjectQueryLoading ? (
                <>
                  <Skeleton height={'16px'} />
                  <Skeleton width={'50%'} height={'16px'} />
                  <Skeleton height={'32px'} />
                  <Skeleton height={'16px'} />
                  <Skeleton height={'32px'} />
                </>
              ) : (
                <>
                  <Text size="small">
                    This action can not be undone. This will permanently delete
                    project <b>{project?.title}</b>.
                  </Text>

                  <InputField
                    label="Please type name of the project to confirm."
                    size="large"
                    error={errors.name && String(errors.name?.message)}
                    {...register('name')}
                    placeholder="Provide project name"
                  />

                  <Flex gap="small">
                    <Checkbox
                      checked={isAcknowledged}
                      onCheckedChange={v => setIsAcknowledged(v === true)}
                      data-test-id="frontier-sdk-delete-project-checkbox"
                    />
                    <Text size="small">
                      I understand that all of the project data will be deleted
                      and want to proceed.
                    </Text>
                  </Flex>
                  <Button
                    variant="solid"
                    color="danger"
                    type="submit"
                    disabled={!name || !isAcknowledged}
                    style={{ width: '100%' }}
                    data-test-id="frontier-sdk-delete-project-btn"
                    loading={isSubmitting}
                    loaderText="Deleting..."
                  >
                    Delete this project
                  </Button>
                </>
              )}
            </Flex>
          </form>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
