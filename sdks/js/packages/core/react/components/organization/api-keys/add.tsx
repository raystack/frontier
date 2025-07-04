import {
  Button,
  Flex,
  Text,
  toast,
  Image,
  Skeleton,
  Dialog,
  InputField,
  Select,
  Label
} from '@raystack/apsara/v1';
import { useNavigate } from '@tanstack/react-router';
import { Controller, useForm } from 'react-hook-form';
import { useFrontier } from '~/react/contexts/FrontierContext';
import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { useCallback, useEffect, useState } from 'react';
import { V1Beta1CreatePolicyForProjectBody, V1Beta1Project } from '~/src';
import { PERMISSIONS } from '~/utils';
import cross from '~/react/assets/cross.svg';
import styles from './styles.module.css';
import { handleSelectValueChange } from '~/react/utils';

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
    register,
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
      <Dialog.Content
        overlayClassName={styles.overlay}
        className={styles.addDialogContent}
      >
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Header>
            <Flex justify="between" align="center" style={{ width: '100%' }}>
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
          </Dialog.Header>

          <Dialog.Body>
            <Flex direction="column" gap={5}>
              <Text>
                Create a dedicated service account to facilitate secure API
                interactions on behalf of the organization.
              </Text>
              {isLoading ? (
                <Skeleton height={'25px'} />
              ) : (
                <InputField
                  label="Name"
                  {...register('title')}
                  size="medium"
                  placeholder="Provide service account name"
                  error={errors.title && String(errors.title?.message)}
                />
              )}
              <Flex direction="column" gap={2}>
                <Label>Project</Label>
                {isLoading ? (
                  <Skeleton height={'25px'} />
                ) : (
                  <Controller
                    render={({ field }) => {
                      const { ref, onChange, ...rest } = field;
                      return (
                        <Select
                          {...rest}
                          onValueChange={handleSelectValueChange(onChange)}
                        >
                          <Select.Trigger ref={ref}>
                            <Select.Value placeholder="Select a project" />
                          </Select.Trigger>
                          <Select.Content style={{ width: '100% !important' }}>
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
                <Text size="mini" variant="danger">
                  {errors.project_id && String(errors.project_id?.message)}
                </Text>
              </Flex>
            </Flex>
          </Dialog.Body>

          <Dialog.Footer>
            <Flex justify="end">
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
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
