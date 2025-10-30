import { yupResolver } from '@hookform/resolvers/yup';
import {
  Button,
  Separator,
  toast,
  Tooltip,
  Skeleton,
  Box,
  Flex,
  InputField
} from '@raystack/apsara';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { createConnectQueryKey, useMutation, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { FrontierServiceQueries, UpdateOrganizationRequestSchema, Organization } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { AuthTooltipMessage } from '~/react/utils';
import { AvatarUpload } from '../../avatar-upload';
import { getInitials } from '~/utils';
import { useTerminology } from '~/react/hooks/useTerminology';
import styles from './general.module.css';

const generalSchema = yup
  .object({
    avatar: yup.string().optional(),
    title: yup.string().required('Name is a required field'),
    name: yup.string().required('URL is a required field')
  })
  .required();

type FormData = yup.InferType<typeof generalSchema>;

interface GeneralOrganizationProps {
  organization?: Organization;
  isLoading?: boolean;
  canUpdateWorkspace?: boolean;
}

export const GeneralOrganization = ({
  organization,
  isLoading,
  canUpdateWorkspace = false
}: GeneralOrganizationProps) => {
  const { setActiveOrganization } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();
  const { mutateAsync: updateOrganization } = useMutation(
    FrontierServiceQueries.updateOrganization,
    {
      onSuccess: (data) => {
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
        toast.success(`Updated ${t.organization({ case: 'lower' })}`);
      },
      onError: (error: Error) => {
        toast.error('Something went wrong', {
          description: error?.message || 'Failed to update'
        });
      },
    }
  );
  const t = useTerminology();
  const {
    reset,
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(generalSchema)
  });

  const URL_PREFIX = window?.location?.host + '/';
  
  useEffect(() => {
    reset(organization);
  }, [organization, reset]);

  async function onSubmit(data: FormData) {
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
    } catch (error: unknown) {
      toast.error('Something went wrong', {
        description: error instanceof Error ? error.message : 'Failed to update organization'
      });
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Flex direction="column" gap={9} style={{ maxWidth: '320px' }}>
        {isLoading ? (
          <Flex gap={5} direction="column" style={{ width: '100%' }}>
            <Skeleton width="80px" height="80px" borderRadius="50%" />
            <Skeleton height="16px" width="100%" />
          </Flex>
        ) : (
          <AvatarUpload
            value={watch('avatar')}
            onChange={(value) => setValue('avatar', value)}
            subText={`Pick a logo for your ${t.organization({
              case: 'lower'
            })}.`}
            initials={getInitials(
              organization?.title || organization?.name
            )}
            disabled={!canUpdateWorkspace}
            data-test-id="frontier-sdk-avatar-upload"
          />
        )}
      </Flex>
      <Separator className={styles.separator} />
      <Flex direction="column" gap={9} style={{ maxWidth: '320px' }}>
        <Box style={{ padding: 'var(--rs-space-2) 0' }}>
          {isLoading ? (
            <>
              <Skeleton height={'16px'} />
              <Skeleton height={'32px'} />
            </>
          ) : (
            <InputField
              label={`${t.organization({ case: 'capital' })} name`}
              size="large"
              error={errors.title && String(errors.title?.message)}
              defaultValue={organization?.title || ''}
              disabled={!canUpdateWorkspace}
              placeholder={`Provide ${t.organization({ case: 'lower' })} name`}
              {...register('title')}
            />
          )}
        </Box>
        <Box style={{ padding: 'var(--rs-space-2) 0' }}>
          {isLoading ? (
            <>
              <Skeleton height={'16px'} />
              <Skeleton height={'32px'} />
            </>
          ) : (
            <InputField
              label={`${t.organization({ case: 'capital' })} URL`}
              size="large"
              error={errors.name && String(errors.name?.message)}
              defaultValue={organization?.name || ''}
              disabled
              prefix={URL_PREFIX}
              placeholder={`Provide ${t.organization({ case: 'lower' })} URL`}
              {...register('name')}
            />
          )}
        </Box>
        {isLoading ? (
          <Skeleton height={'32px'} width={'64px'} />
        ) : (
          <Tooltip message={AuthTooltipMessage} disabled={canUpdateWorkspace}>
            <Button
              type="submit"
              style={{ width: 'fit-content' }}
              disabled={isLoading || isSubmitting || !canUpdateWorkspace}
              data-test-id="frontier-sdk-update-organization-btn"
              loading={isSubmitting}
              loaderText="Updating..."
            >
              Update
            </Button>
          </Tooltip>
        )}
      </Flex>
    </form>
  );
};
