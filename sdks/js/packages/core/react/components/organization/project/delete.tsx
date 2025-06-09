import { InputField, TextField } from '@raystack/apsara';
import {
  Button,
  Checkbox,
  Separator,
  toast,
  Skeleton,
  Image,
  Text,
  Flex,
  Dialog
} from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Project } from '~/src';
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
    control,
    setError,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(projectSchema)
  });
  let { projectId } = useParams({ from: '/projects/$projectId/delete' });
  const navigate = useNavigate({ from: '/projects/$projectId/delete' });
  const { client, activeOrganization: organization } = useFrontier();
  const [isProjectLoading, setIsProjectLoading] = useState(false);
  const [project, setProject] = useState<V1Beta1Project>();
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  useEffect(() => {
    async function getProjectDetails() {
      if (!projectId) return;
      try {
        setIsProjectLoading(true);
        const {
          // @ts-ignore
          data: { project }
        } = await client?.frontierServiceGetProject(projectId);
        setProject(project);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      } finally {
        setIsProjectLoading(false);
      }
    }
    getProjectDetails();
  }, [client, projectId]);

  async function onSubmit(data: any) {
    if (!organization?.id) return;
    if (!projectId) return;
    if (!client) return;

    if (data.name !== project?.name)
      return setError('name', { message: 'project name is not same' });

    try {
      await client.frontierServiceDeleteProject(projectId);
      toast.success('project deleted');
      navigate({ to: '/projects' });
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  const name = watch('name', '');
  return (
    <Dialog open={true}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassName={styles.overlay}
      >
        <Dialog.Header>
          <Flex justify="between" style={{ padding: '16px 24px' }}>
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
          <Separator />
        </Dialog.Header>

        <Dialog.Body>
          <form onSubmit={handleSubmit(onSubmit)}>
            <Flex direction="column" gap={5} style={{ padding: '24px 32px' }}>
              {isProjectLoading ? (
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

                  <InputField label="Please type name of the project to confirm.">
                    <Controller
                      render={({ field }) => (
                        <TextField
                          {...field}
                          // @ts-ignore
                          size="medium"
                          placeholder="Provide project name"
                        />
                      )}
                      control={control}
                      name="name"
                    />

                    <Text size="mini" variant="danger">
                      {errors.name && String(errors.name?.message)}
                    </Text>
                  </InputField>
                  <Flex gap={3}>
                    <Checkbox
                      checked={isAcknowledged}
                      onCheckedChange={v => setIsAcknowledged(v === true)}
                      data-test-id="frontier-sdk-delete-project-checkbox"
                    />
                    <Text size="small">
                      I acknowledge I understand that all of the project data
                      will be deleted and want to proceed.
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
