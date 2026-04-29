'use client';

import { useCallback, useMemo } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { orderBy } from 'lodash';
import { create } from '@bufbuild/protobuf';
import { useQueryClient } from '@tanstack/react-query';
import {
  useMutation,
  useQuery,
  createConnectQueryKey,
  useTransport
} from '@connectrpc/connect-query';
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
import {
  Button,
  Text,
  Dialog,
  Flex,
  InputField,
  Label,
  Select,
  Skeleton,
  toastManager
} from '@raystack/apsara-v1';
import { PERMISSIONS } from '~/utils';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useTerminology } from '~/react/hooks/useTerminology';
import { cacheFreshServiceUserToken } from '../hooks/useServiceUserTokens';

const DEFAULT_KEY_NAME = 'Initial Generated Key';

const serviceAccountSchema = yup
  .object({
    title: yup.string().required('Name is a required field'),
    project_ids: yup
      .array()
      .of(yup.string().required())
      .min(1, 'At least one project is required')
      .required('Project is a required field')
  })
  .required();

type FormData = yup.InferType<typeof serviceAccountSchema>;

export interface AddServiceAccountDialogProps {
  handle: ReturnType<typeof Dialog.createHandle>;
  onCreated?: (serviceUserId: string) => void;
}

export function AddServiceAccountDialog({
  handle,
  onCreated
}: AddServiceAccountDialogProps) {
  const { activeOrganization: organization } = useFrontier();
  const t = useTerminology();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const orgId = organization?.id || '';

  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting, isDirty }
  } = useForm<FormData>({
    resolver: yupResolver(serviceAccountSchema),
    defaultValues: {
      title: '',
      project_ids: []
    }
  });

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      reset();
    }
  };

  const { data: projectsData, isLoading: isProjectsLoading } = useQuery(
    FrontierServiceQueries.listOrganizationProjects,
    create(ListOrganizationProjectsRequestSchema, {
      id: orgId,
      state: '',
      withMemberCount: false
    }),
    {
      enabled: Boolean(orgId),
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

        await Promise.all(
          data.project_ids.map(projectId =>
            createPolicyForProject(
              create(CreatePolicyForProjectRequestSchema, {
                projectId,
                body: create(CreatePolicyForProjectBodySchema, {
                  roleId: PERMISSIONS.RoleProjectOwner,
                  principal
                })
              })
            )
          )
        );

        const tokenResponse = await createServiceUserToken(
          create(CreateServiceUserTokenRequestSchema, {
            orgId,
            id: serviceUserId,
            title: DEFAULT_KEY_NAME
          })
        );

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

        if (tokenResponse.token) {
          cacheFreshServiceUserToken(
            queryClient,
            serviceUserId,
            tokenResponse.token
          );
        }

        toastManager.add({ title: 'Service account created', type: 'success' });
        handle.close();
        reset();
        onCreated?.(serviceUserId);
      } catch (error: unknown) {
        toastManager.add({
          title: 'Something went wrong',
          description: error instanceof Error ? error.message : 'Unknown error',
          type: 'error'
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
      handle,
      reset,
      onCreated
    ]
  );

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      <Dialog.Content width={400}>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Header>
            <Dialog.Title>New Service Account</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Flex direction="column" gap={7}>
              <Text size="small">
                Create a dedicated service account to facilitate secure API
                interactions on behalf of the{' '}
                {t.organization({ case: 'lower' })}.
              </Text>
              {isProjectsLoading ? (
                <Flex direction="column" gap={5}>
                  <Skeleton height="60px" />
                  <Skeleton height="60px" />
                </Flex>
              ) : (
                <>
                  <InputField
                    label="Name"
                    {...register('title')}
                    size="large"
                    placeholder="Provide service account name"
                    error={errors.title && String(errors.title?.message)}
                  />
                  <Flex direction="column" gap={2}>
                    <Label>Project</Label>
                    <Controller
                      name="project_ids"
                      control={control}
                      render={({ field }) => (
                        <Select
                          multiple
                          value={field.value}
                          onValueChange={field.onChange}
                        >
                          <Select.Trigger>
                            <Select.Value placeholder="Select projects" />
                          </Select.Trigger>
                          <Select.Content>
                            {projects.map(project => (
                              <Select.Item
                                value={project.id || ''}
                                key={project.id}
                              >
                                {project.title}
                              </Select.Item>
                            ))}
                          </Select.Content>
                        </Select>
                      )}
                    />
                    {errors.project_ids && (
                      <Text size="mini" variant="danger">
                        {String(errors.project_ids?.message)}
                      </Text>
                    )}
                  </Flex>
                </>
              )}
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
                disabled={isSubmitting || isProjectsLoading || !isDirty}
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
}
