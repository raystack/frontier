import { yupResolver } from '@hookform/resolvers/yup';
import {
  Box,
  Button,
  Flex,
  InputField,
  Separator,
  Text,
  TextField,
  Tooltip
} from '@raystack/apsara';
import React, { forwardRef, useCallback, useEffect, useRef } from 'react';
import { Controller, useForm } from 'react-hook-form';
import Skeleton from 'react-loading-skeleton';
import { toast } from 'sonner';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Organization } from '~/src';
// @ts-ignore
import styles from './general.module.css';
import { AuthTooltipMessage } from '~/react/utils';
import { AvatarUpload } from '../../avatar-upload';
import { getInitials } from '~/utils';

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
      <div onClick={focusChild} className={styles.prefixInput}>
        <Text size={2} style={{ color: 'var(--foreground-muted)' }}>
          {prefix}
        </Text>
        <input {...props} ref={setRef} />
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
  const { client, setActiveOrganization } = useFrontier();
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
    if (!client) return;
    if (!organization?.id) return;

    // This API call can be moved to sdks/js/packages/core/react/components/organization/general/index.tsx
    try {
      const resp = await client.frontierServiceUpdateOrganization(
        organization?.id,
        data
      );
      if (resp.data?.organization) {
        setActiveOrganization(resp.data?.organization);
      }
      toast.success('Updated organization');
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Flex direction="column" gap="large" style={{ maxWidth: '320px' }}>
        {isLoading ? (
          <Flex gap={'medium'} direction={'column'} style={{ width: '100%' }}>
            <Skeleton style={{ width: '80px', height: '80px' }} circle />
            <Skeleton style={{ height: '16px', width: '100%' }} />
          </Flex>
        ) : (
          <Controller
            render={({ field }) => (
              <AvatarUpload
                {...field}
                subText="Pick a logo for your organization."
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
      <Separator style={{ margin: '32px 0' }} />
      <Flex direction="column" gap="large" style={{ maxWidth: '320px' }}>
        <Box style={{ padding: 'var(--pd-4) 0' }}>
          {isLoading ? (
            <>
              <Skeleton height={'16px'} />
              <Skeleton height={'32px'} />
            </>
          ) : (
            <InputField label="Organization name">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide organization name"
                  />
                )}
                defaultValue={organization?.title}
                control={control}
                disabled={!canUpdateWorkspace}
                name="title"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.title && String(errors.title?.message)}
              </Text>
            </InputField>
          )}
        </Box>
        <Box style={{ padding: 'var(--pd-4) 0' }}>
          {isLoading ? (
            <>
              <Skeleton height={'16px'} />
              <Skeleton height={'32px'} />
            </>
          ) : (
            <InputField label="Organization URL">
              <Controller
                render={({ field }) => (
                  <PrefixInput
                    prefix={URL_PREFIX}
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide organization URL"
                    disabled
                  />
                )}
                defaultValue={organization?.name}
                control={control}
                name="name"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.name && String(errors.name?.message)}
              </Text>
            </InputField>
          )}
        </Box>
        {isLoading ? (
          <Skeleton height={'32px'} width={'64px'} />
        ) : (
          <Tooltip message={AuthTooltipMessage} disabled={canUpdateWorkspace}>
            <Button
              size="medium"
              variant="primary"
              type="submit"
              style={{ width: 'fit-content' }}
              disabled={isLoading || isSubmitting || !canUpdateWorkspace}
              data-test-id="frontier-sdk-update-organization-btn"
            >
              {isSubmitting ? 'Updating...' : 'Update'}
            </Button>
          </Tooltip>
        )}
      </Flex>
    </form>
  );
};
