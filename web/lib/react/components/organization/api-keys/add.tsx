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
} from '@raystack/apsara';
import { useNavigate } from '@tanstack/react-router';
import { Controller, useForm } from 'react-hook-form';
import { useFrontier } from '~/react/contexts/FrontierContext';
import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { useCallback, useMemo } from 'react';
import { orderBy } from 'lodash';
import {
  FrontierServiceQueries,
  CreateServiceUserRequestSchema,
  CreatePolicyForProjectRequestSchema,
  CreateServiceUserTokenRequestSchema,
  ListOrganizationServiceUsersRequestSchema,
  ListOrganizationProjectsRequestSchema,
  ListServiceUserTokensRequestSchema,
  ListServiceUserTokensResponseSchema,
  ServiceUserRequestBodySchema,
  CreatePolicyForProjectBodySchema
} from '@raystack/proton/frontier';
import { PERMISSIONS } from '~/utils';
import { useQueryClient } from '@tanstack/react-query';
import {
  useMutation,
  useQuery,
  createConnectQueryKey,
  useTransport
} from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import cross from '~/react/assets/cross.svg';
import styles from './styles.module.css';
import { handleSelectValueChange } from '~/react/utils';
import { useTerminology } from '~/react/hooks/useTerminology';

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
  const { activeOrganization: organization } = useFrontier();
  const t = useTerminology();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const {
    register,
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(serviceAccountSchema)
  });

  const orgId = organization?.id || '';

  const { data: projectsData, isLoading: isProjectsLoading } = useQuery(
    FrontierServiceQueries.listOrganizationProjects,
    create(ListOrganizationProjectsRequestSchema, {
      id: orgId,
      state: '',
      withMemberCount: false
    }),
    {
      enabled: Boolean(orgId)
    }
  );

  const projects = useMemo(() => {
    const list = projectsData?.projects ?? [];
    return orderBy(list, ['title'], ['asc']);
  }, [projectsData]);

  const { mutateAsync: createServiceUser } = useMutation(
    FrontierServiceQueries.createServiceUser
  );

  const { mutateAsync: createPolicyForProject } = useMutation(
    FrontierServiceQueries.createPolicyForProject
  );

  const { mutateAsync: createServiceUserToken } = useMutation(
    FrontierServiceQueries.createServiceUserToken
  );

  const onSubmit = useCallback(
    async (data: FormData) => {
      if (!orgId) return;

      try {
        const serviceUserResponse = await createServiceUser(
          create(CreateServiceUserRequestSchema, {
            orgId,
            body: create(ServiceUserRequestBodySchema, {
              title: data.title
            })
          })
        );

        const serviceUserId = serviceUserResponse.serviceuser?.id;
        if (!serviceUserId) return;

        const principal = `${PERMISSIONS.ServiceUserPrincipal}:${serviceUserId}`;

        await createPolicyForProject(
          create(CreatePolicyForProjectRequestSchema, {
            projectId: data.project_id,
            body: create(CreatePolicyForProjectBodySchema, {
              roleId: PERMISSIONS.RoleProjectOwner,
              principal
            })
          })
        );

        const tokenResponse = await createServiceUserToken(
          create(CreateServiceUserTokenRequestSchema, {
            orgId,
            id: serviceUserId,
            title: DEFAULT_KEY_NAME
          })
        );

        // Invalidate service users query
        await queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.listOrganizationServiceUsers,
            transport,
            input: create(ListOrganizationServiceUsersRequestSchema, {
              id: orgId
            }),
            cardinality: 'finite'
          })
        });

        // Seed listServiceUserTokens cache with the new token
        const listTokensQueryKey = createConnectQueryKey({
          schema: FrontierServiceQueries.listServiceUserTokens,
          transport,
          input: create(ListServiceUserTokensRequestSchema, {
            id: serviceUserId,
            orgId
          }),
          cardinality: 'finite'
        });

        queryClient.setQueryData(
          listTokensQueryKey,
          create(ListServiceUserTokensResponseSchema, {
            tokens: tokenResponse.token ? [tokenResponse.token] : []
          })
        );

        toast.success('Service user created');

        navigate({
          to: '/api-keys/$id',
          params: { id: serviceUserId }
        });
      } catch (error: unknown) {
        toast.error('Something went wrong', {
          description: error instanceof Error ? error.message : 'Unknown error'
        });
      }
    },
    [
      orgId,
      createServiceUser,
      createPolicyForProject,
      createServiceUserToken,
      queryClient,
      transport,
      navigate
    ]
  );

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
                interactions on behalf of the{' '}
                {t.organization({ case: 'lower' })}.
              </Text>
              {isProjectsLoading ? (
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
                {isProjectsLoading ? (
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
                loading={isSubmitting}
                disabled={isSubmitting || isProjectsLoading}
                loaderText="Creating..."
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
