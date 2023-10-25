import {
  Button,
  Dialog,
  Flex,
  Image,
  InputField,
  Separator,
  Text,
  TextField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { toast } from 'sonner';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Project } from '~/src';
import styles from '../organization.module.css';
import Skeleton from 'react-loading-skeleton';

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
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} style={{ fontWeight: '500' }}>
            Verify project deletion
          </Text>
          <Image
            alt="cross"
            // @ts-ignore
            src={cross}
            onClick={() =>
              navigate({ to: '/projects/$projectId', params: { projectId } })
            }
            style={{ cursor: 'pointer' }}
          />
        </Flex>
        <Separator />
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex
            direction="column"
            gap="medium"
            style={{ padding: '24px 32px' }}
          >
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
                <Text size={2}>
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

                  <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                    {errors.name && String(errors.name?.message)}
                  </Text>
                </InputField>
                <Flex>
                  <Text size={2}>
                    I acknowledge I understand that all of the project data will
                    be deleted and want to proceed.
                  </Text>
                </Flex>
                <Button
                  variant="danger"
                  size="medium"
                  type="submit"
                  disabled={!name}
                  style={{ width: '100%' }}
                >
                  {isSubmitting ? 'deleting...' : 'Delete this project'}
                </Button>
              </>
            )}
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
