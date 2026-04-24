'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { orderBy } from 'lodash';
import { create } from '@bufbuild/protobuf';
import { timestampFromDate } from '@bufbuild/protobuf/wkt';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import dayjs from 'dayjs';
import {
  FrontierServiceQueries,
  CreateCurrentUserPATRequestSchema,
  UpdateCurrentUserPATRequestSchema,
  CheckCurrentUserPATTitleRequestSchema,
  ListRolesForPATRequestSchema,
  ListOrganizationProjectsRequestSchema
} from '@raystack/proton/frontier';
import type { PAT } from '@raystack/proton/frontier';
import {
  Button,
  Dialog,
  Flex,
  InputField,
  Label,
  Radio,
  Select,
  Skeleton,
  Spinner,
  Text,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_DATE_FORMAT } from '~/react/utils/constants';
import { PERMISSIONS } from '../../../../utils';
import { handleConnectError } from '~/utils/error';

const EXPIRY_OPTIONS = [
  { value: '1w', label: '1 week', amount: 1, unit: 'week' as const },
  { value: '1m', label: '1 month', amount: 1, unit: 'month' as const },
  { value: '3m', label: '3 months', amount: 3, unit: 'month' as const },
  { value: '6m', label: '6 months', amount: 6, unit: 'month' as const },
  { value: '12m', label: '12 months', amount: 12, unit: 'month' as const }
] as const;

const baseFields = {
  title: yup.string().required('Name is required'),
  orgRoleId: yup.string().required('Organization role is required'),
  projectRoleId: yup.string().required('Project role is required'),
  projectIds: yup
    .array()
    .of(yup.string().required())
    .default([])
};

const createPATSchema = yup
  .object({
    ...baseFields,
    expiry: yup.string().required('Expiry date is required')
  })
  .required();

const updatePATSchema = yup
  .object({
    ...baseFields,
    expiry: yup.string().default('')
  })
  .required();

type FormData = yup.InferType<typeof createPATSchema>;

export interface PATFormDialogProps {
  handle: ReturnType<typeof Dialog.createHandle>;
  initialData?: PAT;
  onCreated?: (token: string) => void;
  onUpdated?: () => void;
}

export function PATFormDialog({
  handle,
  initialData,
  onCreated,
  onUpdated
}: PATFormDialogProps) {
  const { activeOrganization: organization, config } = useFrontier();
  const orgId = organization?.id || '';
  const dateFormat = config?.dateFormat || DEFAULT_DATE_FORMAT;

  const isUpdateMode = Boolean(initialData);

  const {
    register,
    control,
    handleSubmit,
    reset,
    getValues,
    watch,
    setValue,
    setError,
    clearErrors,
    formState: { errors, isSubmitting, isDirty }
  } = useForm<FormData>({
    resolver: yupResolver(isUpdateMode ? updatePATSchema : createPATSchema),
    defaultValues: {
      title: '',
      expiry: '',
      orgRoleId: '',
      projectRoleId: '',
      projectIds: []
    }
  });

  // null = unchecked, true = available, false = taken
  const [titleAvailable, setTitleAvailable] = useState<boolean | null>(null);
  const [titleChecking, setTitleChecking] = useState(false);
  const [projectAccess, setProjectAccess] = useState<'all' | 'selective'>(
    'all'
  );

  const handleOpenChange = (open: boolean) => {
    if (open && initialData) {
      const orgScope = initialData.scopes?.find(
        s => s.resourceType === PERMISSIONS.OrganizationNamespace
      );
      const projectScope = initialData.scopes?.find(
        s => s.resourceType === PERMISSIONS.ProjectNamespace
      );
      const validProjectIds = (projectScope?.resourceIds || []).filter(id =>
        projects.some(p => p.id === id)
      );
      setProjectAccess(validProjectIds.length > 0 ? 'selective' : 'all');
      reset({
        title: initialData.title,
        expiry: '',
        orgRoleId: orgScope?.roleId || '',
        projectRoleId: projectScope?.roleId || '',
        projectIds: validProjectIds
      });
      setTitleAvailable(true);
    }
    if (!open) {
      reset();
      setTitleAvailable(null);
      setProjectAccess('all');
    }
  };

  const { data: projectsData, isLoading: isProjectsLoading } = useQuery(
    FrontierServiceQueries.listOrganizationProjects,
    create(ListOrganizationProjectsRequestSchema, {
      id: orgId,
      state: '',
      withMemberCount: false
    }),
    { enabled: Boolean(orgId) }
  );

  const projects = useMemo(() => {
    const list = projectsData?.projects ?? [];
    return orderBy(list, ['title'], ['asc']);
  }, [projectsData]);

  const { data: rolesData, isLoading: isRolesLoading } = useQuery(
    FrontierServiceQueries.listRolesForPAT,
    create(ListRolesForPATRequestSchema, { scopes: [] }),
    { enabled: Boolean(orgId) }
  );

  const { orgRoles, projectRoles } = useMemo(() => {
    const roles = rolesData?.roles ?? [];
    return {
      orgRoles: roles.filter(r =>
        r.scopes?.includes(PERMISSIONS.OrganizationNamespace)
      ),
      projectRoles: roles.filter(r =>
        r.scopes?.includes(PERMISSIONS.ProjectNamespace)
      )
    };
  }, [rolesData]);

  const watchedOrgRoleId = watch('orgRoleId');
  const isOrgAdmin = useMemo(() => {
    const role = orgRoles.find(r => r.id === watchedOrgRoleId);
    if (!role) return false;
    const orgNs = PERMISSIONS.OrganizationNamespace.replace('/', '_');
    const updatePerm = `${orgNs}_${PERMISSIONS.UpdatePermission}`;
    return role.permissions?.some(p => p === updatePerm) ?? false;
  }, [orgRoles, watchedOrgRoleId]);

  useEffect(() => {
    if (isOrgAdmin) {
      setProjectAccess('all');
      setValue('projectIds', []);
    }
  }, [isOrgAdmin, setValue]);

  const { mutateAsync: createPAT } = useMutation(
    FrontierServiceQueries.createCurrentUserPAT
  );

  const { mutateAsync: updatePAT } = useMutation(
    FrontierServiceQueries.updateCurrentUserPAT
  );

  const { mutateAsync: checkTitle } = useMutation(
    FrontierServiceQueries.checkCurrentUserPATTitle
  );

  const handleTitleBlur = useCallback(async () => {
    const title = getValues('title');
    if (!title || !orgId) return;

    // In update mode, skip check if title is unchanged
    if (isUpdateMode && title === initialData?.title) {
      setTitleAvailable(true);
      return;
    }

    setTitleChecking(true);
    try {
      const result = await checkTitle(
        create(CheckCurrentUserPATTitleRequestSchema, { orgId, title })
      );
      setTitleAvailable(result?.available);
    } catch {
      // Ignore check failure — don't block the user
    } finally {
      setTitleChecking(false);
    }
  }, [getValues, orgId, checkTitle, isUpdateMode, initialData]);

  const titleField = register('title');

  const onSubmit = useCallback(
    async (data: FormData) => {
      if (!orgId) return;

      if (
        projectAccess === 'selective' &&
        (!data.projectIds || data.projectIds.length === 0)
      ) {
        setError('projectIds', {
          type: 'manual',
          message: 'At least one project is required'
        });
        return;
      }

      const scopes = [
        {
          roleId: data.orgRoleId,
          resourceType: PERMISSIONS.OrganizationNamespace,
          resourceIds: [] as string[]
        },
        {
          roleId: data.projectRoleId,
          resourceType: PERMISSIONS.ProjectNamespace,
          resourceIds: projectAccess === 'all' ? [] : data.projectIds
        }
      ];

      try {
        if (isUpdateMode && initialData) {
          await updatePAT(
            create(UpdateCurrentUserPATRequestSchema, {
              id: initialData.id,
              title: data.title,
              scopes
            })
          );
          toastManager.add({
            title: 'Personal access token updated',
            type: 'success'
          });
          handle.close();
          reset();
          onUpdated?.();
        } else {
          const option = EXPIRY_OPTIONS.find(o => o.value === data.expiry);
          if (!option) return;
          const expiresAt = timestampFromDate(
            dayjs().add(option.amount, option.unit).toDate()
          );
          const response = await createPAT(
            create(CreateCurrentUserPATRequestSchema, {
              title: data.title,
              orgId,
              scopes,
              expiresAt
            })
          );
          const token = response.pat?.token;
          toastManager.add({
            title: 'Personal access token created',
            type: 'success'
          });
          handle.close();
          reset();
          if (token) onCreated?.(token);
        }
      } catch (error) {
        handleConnectError(error, {
          AlreadyExists: () =>
            toastManager.add({
              title: 'A token with this name already exists',
              type: 'error'
            }),
          Default: err =>
            toastManager.add({
              title: 'Something went wrong',
              description: err.message,
              type: 'error'
            })
        });
      }
    },
    [
      orgId,
      isUpdateMode,
      initialData,
      createPAT,
      updatePAT,
      handle,
      reset,
      onCreated,
      onUpdated,
      projectAccess,
      setError
    ]
  );

  const isDataLoading = isProjectsLoading || isRolesLoading;

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      <Dialog.Content width={400}>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Header>
            <Dialog.Title>
              {isUpdateMode ? 'Update PAT' : 'Create new PAT'}
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Flex direction="column" gap={7}>
              {isDataLoading ? (
                <Flex direction="column" gap={5}>
                  <Skeleton height="60px" />
                  <Skeleton height="60px" />
                  <Skeleton height="60px" />
                  <Skeleton height="60px" />
                  {!isUpdateMode && <Skeleton height="60px" />}
                </Flex>
              ) : (
                <>
                  <InputField
                    label="Name"
                    name={titleField.name}
                    ref={titleField.ref}
                    onChange={e => {
                      titleField.onChange(e);
                      if (
                        isUpdateMode &&
                        e.target.value === initialData?.title
                      ) {
                        setTitleAvailable(true);
                      } else {
                        setTitleAvailable(null);
                      }
                    }}
                    onBlur={async e => {
                      titleField.onBlur(e);
                      await handleTitleBlur();
                    }}
                    size="large"
                    placeholder="Enter token name"
                    trailingIcon={titleChecking ? <Spinner size={2} /> : undefined}
                    error={
                      errors.title
                        ? String(errors.title?.message)
                        : titleAvailable === false
                          ? 'This name is already taken'
                          : undefined
                    }
                  />

                  {!isUpdateMode && (
                    <Flex direction="column" gap={2}>
                      <Label>Expiry date</Label>
                      <Controller
                        name="expiry"
                        control={control}
                        render={({ field }) => (
                          <Select
                            value={field.value}
                            onValueChange={field.onChange}
                          >
                            <Select.Trigger>
                              <Select.Value placeholder="Select expiry" />
                            </Select.Trigger>
                            <Select.Content>
                              {EXPIRY_OPTIONS.map(option => (
                                <Select.Item
                                  key={option.value}
                                  value={option.value}
                                >
                                  {option.label} (Exp:{' '}
                                  {dayjs()
                                    .add(option.amount, option.unit)
                                    .format(dateFormat)}
                                  )
                                </Select.Item>
                              ))}
                            </Select.Content>
                          </Select>
                        )}
                      />
                      {errors.expiry && (
                        <Text size="mini" variant="danger">
                          {String(errors.expiry?.message)}
                        </Text>
                      )}
                    </Flex>
                  )}

                  <Flex direction="column" gap={2}>
                    <Label>Organization Role</Label>
                    <Controller
                      name="orgRoleId"
                      control={control}
                      render={({ field }) => (
                        <Select
                          value={field.value}
                          onValueChange={field.onChange}
                        >
                          <Select.Trigger>
                            <Select.Value placeholder="Select role" />
                          </Select.Trigger>
                          <Select.Content>
                            {orgRoles.map(role => (
                              <Select.Item key={role.id} value={role.id}>
                                {role.title || role.name}
                              </Select.Item>
                            ))}
                          </Select.Content>
                        </Select>
                      )}
                    />
                    {errors.orgRoleId && (
                      <Text size="mini" variant="danger">
                        {String(errors.orgRoleId?.message)}
                      </Text>
                    )}
                  </Flex>

                  <Flex direction="column" gap={2}>
                    <Label>Project Role</Label>
                    <Controller
                      name="projectRoleId"
                      control={control}
                      render={({ field }) => (
                        <Select
                          value={field.value}
                          onValueChange={field.onChange}
                        >
                          <Select.Trigger>
                            <Select.Value placeholder="Select role" />
                          </Select.Trigger>
                          <Select.Content>
                            {projectRoles.map(role => (
                              <Select.Item key={role.id} value={role.id}>
                                {role.title || role.name}
                              </Select.Item>
                            ))}
                          </Select.Content>
                        </Select>
                      )}
                    />
                    {errors.projectRoleId && (
                      <Text size="mini" variant="danger">
                        {String(errors.projectRoleId?.message)}
                      </Text>
                    )}
                  </Flex>

                  <Flex direction="column" gap={5}>
                    <Flex direction="column" gap={4}>
                      <Label>Projects</Label>
                      <Radio.Group
                        value={projectAccess}
                        onValueChange={(val: string) => {
                          const next = val as 'all' | 'selective';
                          setProjectAccess(next);
                          if (next === 'all') {
                            setValue('projectIds', [], {
                              shouldDirty: true
                            });
                          }
                          clearErrors('projectIds');
                        }}
                      >
                        <Flex
                          style={{ gap: 'var(--rs-space-10)' }}
                          align="center"
                        >
                          <Flex gap={3} align="center">
                            <Radio value="all" />
                            <Text size="small" variant="secondary">
                              All
                            </Text>
                          </Flex>
                          <Flex gap={3} align="center">
                            <Radio
                              value="selective"
                              disabled={isOrgAdmin}
                            />
                            <Text size="small" variant="secondary">
                              Selective projects
                            </Text>
                          </Flex>
                        </Flex>
                      </Radio.Group>
                    </Flex>
                    {projectAccess === 'selective' && (
                      <Controller
                        name="projectIds"
                        control={control}
                        render={({ field }) => (
                          <Select
                            multiple
                            value={field.value}
                            onValueChange={(val: string[]) => {
                              field.onChange(val);
                              clearErrors('projectIds');
                            }}
                          >
                            <Select.Trigger>
                              <Select.Value placeholder="Select projects">
                                {field.value.length > 0
                                  ? `${field.value.length} project${field.value.length > 1 ? 's' : ''} selected`
                                  : undefined}
                              </Select.Value>
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
                    )}
                    {errors.projectIds && (
                      <Text size="mini" variant="danger">
                        {String(errors.projectIds?.message)}
                      </Text>
                    )}
                  </Flex>
                </>
              )}
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Flex justify="end" gap={3}>
              <Button
                variant="outline"
                color="neutral"
                size="normal"
                type="button"
                onClick={() => handle.close()}
              >
                Cancel
              </Button>
              <Button
                variant="solid"
                color="accent"
                size="normal"
                type="submit"
                loading={isSubmitting}
                disabled={
                  isSubmitting ||
                  isDataLoading ||
                  !isDirty ||
                  titleChecking ||
                  titleAvailable !== true
                }
                loaderText={isUpdateMode ? 'Updating...' : 'Creating...'}
                data-test-id="frontier-sdk-pat-form-submit-btn"
              >
                {isUpdateMode ? 'Update' : 'Create'}
              </Button>
            </Flex>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
}
