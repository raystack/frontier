import { yupResolver } from '@hookform/resolvers/yup';
import {
  Button,
  Separator,
  toast,
  Tooltip,
  Skeleton,
  Box,
  Text,
  Flex,
  InputField
} from '@raystack/apsara';
import React, { forwardRef, useCallback, useEffect, useRef } from 'react';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries, UpdateOrganizationRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { V1Beta1Organization } from '~/src';
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

interface PrefixInputProps {
  prefix: string;
}

const PrefixInput = forwardRef<HTMLInputElement, PrefixInputProps>(
  function PrefixInput({ prefix, ...props }, ref) {
    const childRef = useRef<HTMLInputElement | null>(null);

    const focusChild = () => {
      childRef?.current && childRef?.current?.focus();
    };

    const setRef = useCallback(
      (node: HTMLInputElement) => {
        childRef.current = node;
        if (ref != null) {
          (ref as React.MutableRefObject<HTMLInputElement | null>).current =
            node;
        }
      },
      [ref]
    );

    return (
      <div onClick={focusChild} className={styles.prefixInput} data-test-id="frontier-sdk-prefix-input">
        <Text size="small" variant="secondary">
          {prefix}
        </Text>
        <input {...props} ref={setRef} data-test-id="frontier-sdk-prefix-input" />
      </div>
    );
  }
);

export const GeneralOrganization = ({
  organization,
  isLoading,
  canUpdateWorkspace = false
}: {
  organization?: V1Beta1Organization;
  isLoading?: boolean;
  canUpdateWorkspace?: boolean;
}) => {
  const { setActiveOrganization } = useFrontier();
  const { mutateAsync: updateOrganization } = useMutation(
    FrontierServiceQueries.updateOrganization,
  );
  const t = useTerminology();
  const {
    reset,
    control,
    register,
    handleSubmit,
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
          name: organization.name,
          avatar: data.avatar
        }
      });
      const { organization: updated } = await updateOrganization(req);
      if (updated) {
        setActiveOrganization(updated as any);
      }
      toast.success(`Updated ${t.organization({ case: 'lower' })}`);
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message || 'Failed to update'
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
          <Controller
            render={({ field }) => (
              <AvatarUpload
                {...field}
                subText={`Pick a logo for your ${t.organization({
                  case: 'lower'
                })}.`}
                initials={getInitials(
                  organization?.title || organization?.name
                )}
                disabled={!canUpdateWorkspace}
              />
            )}
            control={control}
            name="avatar"
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
