'use client';

import { useEffect, useMemo, useState } from 'react';
import { yupResolver } from '@hookform/resolvers/yup';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { create } from '@bufbuild/protobuf';
import {
  createConnectQueryKey,
  useMutation,
  useTransport
} from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import {
  FrontierServiceQueries,
  UpdateOrganizationRequestSchema
} from '@raystack/proton/frontier';
import {
  Button,
  Tooltip,
  Skeleton,
  Text,
  Flex,
  InputField,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../contexts/FrontierContext';
import { usePermissions } from '../../hooks/usePermissions';
import { useTerminology } from '../../hooks/useTerminology';
import { PERMISSIONS, shouldShowComponent } from '../../../utils';
import { AuthTooltipMessage } from '../../utils';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { ImageUpload } from '../../components/image-upload';
import { DeleteOrganizationDialog } from './components/delete-organization-dialog';
import styles from './general-view.module.css';
import { handleConnectError } from '~/utils/error';

const generalSchema = yup
  .object({
    avatar: yup.string().optional(),
    title: yup.string().required('Name is a required field'),
    name: yup.string().required('URL is a required field')
  })
  .required();

type FormData = yup.InferType<typeof generalSchema>;

export interface GeneralViewProps {
  onDeleteSuccess?: () => void;
  urlPrefix?: string;
}

export function GeneralView({ onDeleteSuccess, urlPrefix }: GeneralViewProps = {}) {
  const t = useTerminology();
  const {
    activeOrganization: organization,
    isActiveOrganizationLoading,
    setActiveOrganization
  } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const resource = `app/organization:${organization?.id}`;

  const listOfPermissionsToCheck = useMemo(() => {
    return [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource: resource
      },
      {
        permission: PERMISSIONS.DeletePermission,
        resource: resource
      }
    ];
  }, [resource]);

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!organization?.id
  );

  const { canUpdateWorkspace, canDeleteWorkspace } = useMemo(() => {
    return {
      canUpdateWorkspace: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      ),
      canDeleteWorkspace: shouldShowComponent(
        permissions,
        `${PERMISSIONS.DeletePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const isLoading = !organization?.id || isActiveOrganizationLoading || isPermissionsFetching;

  // Update organization form
  const { mutateAsync: updateOrganization } = useMutation(
    FrontierServiceQueries.updateOrganization,
    {
      onSuccess: data => {
        if (data.organization) {
          setActiveOrganization(data.organization);
          queryClient.invalidateQueries({
            queryKey: createConnectQueryKey({
              schema: FrontierServiceQueries.getOrganization,
              transport,
              input: { id: organization?.id || '' },
              cardinality: 'finite'
            })
          });
        }
        toastManager.add({
          title: `Updated ${t.organization({ case: 'lower' })}`,
          type: 'success'
        });
      },
    }
  );

  const {
    reset,
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isSubmitting, isDirty }
  } = useForm({
    resolver: yupResolver(generalSchema)
  });

  const URL_PREFIX = urlPrefix ?? window?.location?.host + '/';

  useEffect(() => {
    reset(organization);
  }, [organization, reset]);

  async function onUpdateSubmit(data: FormData) {
    if (!organization?.id) return;

    try {
      const req = create(UpdateOrganizationRequestSchema, {
        id: organization.id,
        body: {
          title: data.title,
          name: data.name,
          avatar: data.avatar
        }
      });
      await updateOrganization(req);
    } catch (error) {
      handleConnectError(error, {
        AlreadyExists: () => toastManager.add({ title: `${t.organization({ case: 'capital' })} already exists`, type: 'error' }),
        InvalidArgument: (err) => toastManager.add({ title: 'Invalid input', description: err.message, type: 'error' }),
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
        NotFound: (err) => toastManager.add({ title: 'Not found', description: err.message, type: 'error' }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
      });
    }
  }

  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  const orgLabel = t.organization({ case: 'capital' });
  const orgLabelLower = t.organization({ case: 'lower' });

  return (
    <ViewContainer>
      <ViewHeader
        title="General"
        description={`Basic configurations for the ${orgLabelLower}`}
      />

      <form onSubmit={handleSubmit(onUpdateSubmit)}>
        <Flex direction="column">
          {/* Logo section */}
          <Flex direction="column" gap={5} className={styles.section}>
            {isLoading ? (
              <Flex gap={5} direction="column">
                <Skeleton width="72px" height="72px" />
                <Skeleton height="20px" width="50%" />
              </Flex>
            ) : (
              <ImageUpload
                value={watch('avatar')}
                onChange={(value: string) => setValue('avatar', value, { shouldDirty: true })}
                description={`Pick a logo for your ${orgLabelLower}. Max size: 5 Mb`}
                disabled={!canUpdateWorkspace}
                data-test-id="frontier-sdk-avatar-upload"
              />
            )}
          </Flex>

          {/* Form section */}
          <Flex direction="column" gap={7} className={styles.section}>
            <Flex
              direction="column"
              gap={9}
              className={styles.formFields}
            >
              {isLoading ? (
                <>
                  <Skeleton height="58px" />
                  <Skeleton height="58px" />
                </>
              ) : (
                <>
                  <InputField
                    label={`${orgLabel} name`}
                    size="large"
                    error={errors.title && String(errors.title?.message)}
                    defaultValue={organization?.title || ''}
                    disabled={!canUpdateWorkspace}
                    placeholder={`Provide ${orgLabelLower} name`}
                    {...register('title')}
                  />
                  <InputField
                    label={`${orgLabel} URL`}
                    size="large"
                    error={errors.name && String(errors.name?.message)}
                    defaultValue={organization?.name || ''}
                    disabled
                    prefix={URL_PREFIX}
                    placeholder={`Provide ${orgLabelLower} URL`}
                    {...register('name')}
                  />
                </>
              )}
            </Flex>
            {isLoading ? (
              <Skeleton height="32px" width="64px" />
            ) : (
              <Tooltip>
                <Tooltip.Trigger
                  disabled={canUpdateWorkspace}
                  render={<span className={styles.fitContent} />}
                >
                  <Button
                    type="submit"
                    variant="solid"
                    color="accent"
                    disabled={
                      isLoading || isSubmitting || !canUpdateWorkspace || !isDirty
                    }
                    data-test-id="frontier-sdk-update-organization-btn"
                    loading={isSubmitting}
                    loaderText="Updating..."
                  >
                    Update
                  </Button>
                </Tooltip.Trigger>
                {!canUpdateWorkspace && (
                  <Tooltip.Content>{AuthTooltipMessage}</Tooltip.Content>
                )}
              </Tooltip>
            )}
          </Flex>

          {/* Delete section */}
          <Flex direction="column" gap={5} className={styles.section}>
            {isLoading ? (
              <>
                <Skeleton height="20px" width="50%" />
                <Skeleton height="32px" width="120px" />
              </>
            ) : (
              <>
                <Text size="regular" variant="secondary">
                  If you want to permanently delete this {orgLabelLower} and all
                  of its data.
                </Text>
                <Tooltip>
                  <Tooltip.Trigger
                    disabled={canDeleteWorkspace}
                    render={<span className={styles.fitContent} />}
                  >
                    <Button
                      variant="solid"
                      color="danger"
                      onClick={() => setShowDeleteDialog(true)}
                      disabled={!canDeleteWorkspace}
                      data-test-id="frontier-sdk-delete-organization-btn"
                    >
                      Delete {orgLabelLower}
                    </Button>
                  </Tooltip.Trigger>
                  {!canDeleteWorkspace && (
                    <Tooltip.Content>{AuthTooltipMessage}</Tooltip.Content>
                  )}
                </Tooltip>
              </>
            )}
          </Flex>
        </Flex>
      </form>
      {canDeleteWorkspace && (
        <DeleteOrganizationDialog
          open={showDeleteDialog}
          onOpenChange={setShowDeleteDialog}
          onDeleteSuccess={onDeleteSuccess}
        />
      )}
    </ViewContainer>
  );
}
