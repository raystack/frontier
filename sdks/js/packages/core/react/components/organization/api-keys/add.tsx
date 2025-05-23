import {
  Dialog,
  InputField,
  Select,
  TextField
} from '@raystack/apsara';
import styles from './styles.module.css';
import { Button, Flex, Text, toast, Separator, Image, Skeleton } from '@raystack/apsara/v1';
import { useNavigate } from '@tanstack/react-router';
import { Controller, useForm } from 'react-hook-form';
import { useFrontier } from '~/react/contexts/FrontierContext';
import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { useCallback, useEffect, useState } from 'react';
import { V1Beta1CreatePolicyForProjectBody, V1Beta1Project } from '~/src';
import { PERMISSIONS } from '~/utils';
import cross from '~/react/assets/cross.svg';

const DEFAULT_KEY_NAME = 'Initial Generated Key';

const serviceAccountSchema = yup
  .object({
    title: yup.string().required('Name is a required field'),
    project_id: yup.string().required('Project is a required field')
  })
  .required();

type FormData = yup.InferType<typeof serviceAccountSchema>;

export const AddServiceAccount = () => {
  const navigate = useNavigate({ from: '/api-keys/add' });
  const { client, activeOrganization: organization } = useFrontier();

  const [projects, setProjects] = useState<V1Beta1Project[]>([]);

  const [isProjectsLoading, setIsProjectsLoading] = useState(false);

  const {
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(serviceAccountSchema)
  });

  const orgId = organization?.id || '';

  const onSubmit = useCallback(
    async (data: FormData) => {
      if (!client || !orgId) return;
      if (!orgId) return;

      try {
        const {
          data: { serviceuser }
        } = await client.frontierServiceCreateServiceUser(orgId, {
          body: data
        });

        if (serviceuser?.id) {
          const principal = `${PERMISSIONS.ServiceUserPrincipal}:${serviceuser?.id}`;

          const policy: V1Beta1CreatePolicyForProjectBody = {
            role_id: PERMISSIONS.RoleProjectViewer,
            principal
          };
          await client?.frontierServiceCreatePolicyForProject(
            data?.project_id,
            policy
          );

          const {
            data: { token }
          } = await client.frontierServiceCreateServiceUserToken(
            orgId,
            serviceuser?.id,
            { title: DEFAULT_KEY_NAME }
          );

          toast.success('Service user created');

          navigate({
            to: '/api-keys/$id',
            params: { id: serviceuser?.id ?? '' },
            state: {
              token: token
            }
          });
        }
      } catch (error: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, navigate, orgId]
  );

  useEffect(() => {
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
    }
  }, [client, orgId]);

  const isDisabled = isSubmitting;

  const isLoading = isProjectsLoading;

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        overlayClassname={styles.overlay}
        className={styles.addDialogContent}
      >
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex justify="between" className={styles.addDialogForm}>
            <Text size={6} weight={500}>
              New Service Account
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={() => navigate({ to: '/api-keys' })}
              data-test-id="frontier-sdk-new-service-account-close-btn"
            />
          </Flex>
          <Separator />

          <Flex
            direction="column"
            gap="medium"
            className={styles.addDialogFormContent}
          >
            <Text>
              Create a dedicated service account to facilitate secure API
              interactions on behalf of the organization.
            </Text>

            <InputField label="Name">
              {isLoading ? (
                <Skeleton height={'25px'} />
              ) : (
                <Controller
                  render={({ field }) => (
                    <TextField
                      {...field}
                      size="medium"
                      placeholder="Provide service account name"
                    />
                  )}
                  name="title"
                  control={control}
                />
              )}
              <Text size={1} variant="danger">
                {errors.title && String(errors.title?.message)}
              </Text>
            </InputField>
            <InputField label="Project">
              {isLoading ? (
                <Skeleton height={'25px'} />
              ) : (
                <Controller
                  render={({ field }) => {
                    const { ref, onChange, ...rest } = field;
                    return (
                      <Select {...rest} onValueChange={onChange}>
                        <Select.Trigger ref={ref}>
                          <Select.Value placeholder="Select a project" />
                        </Select.Trigger>
                        <Select.Content
                          style={{ width: '100% !important', zIndex: 65 }}
                        >
                          <Select.Viewport style={{ maxHeight: '300px' }}>
                            {projects.map(project => (
                              <Select.Item
                                value={project.id || ''}
                                key={project.id}
                              >
                                {project.title}
                              </Select.Item>
                            ))}
                          </Select.Viewport>
                        </Select.Content>
                      </Select>
                    );
                  }}
                  name="project_id"
                  control={control}
                />
              )}
              <Text size={1} variant="danger">
                {errors.project_id && String(errors.project_id?.message)}
              </Text>
            </InputField>
          </Flex>
          <Separator />
          <Flex justify="end" className={styles.addDialogFormBtnWrapper}>
            <Button
              variant="solid"
              color="accent"
              size="normal"
              type="submit"
              data-test-id="frontier-sdk-add-service-account-btn"
              loading={isSubmitting || isLoading}
              disabled={isDisabled || isLoading}
              loaderText={isLoading ? 'Loading...' : 'Creating...'}
            >
              Create
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
