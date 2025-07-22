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
} from '@raystack/apsara/v1';
import React, { forwardRef, useCallback, useEffect, useRef } from 'react';
import { Controller, useForm } from 'react-hook-form';
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
        <Text size="small" variant="secondary">
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
      <Flex direction="column" gap={9} style={{ maxWidth: '320px' }}>
        {isLoading ? (
          <Flex gap={5} direction="column" style={{ width: '100%' }}>
            <Skeleton
              width="80px"
              height="80px"
              borderRadius="50%"
            />
            <Skeleton height="16px" width="100%" />
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
              label="Organization name"
              size="large"
              error={errors.title && String(errors.title?.message)}
              defaultValue={organization?.title || ''}
              disabled={!canUpdateWorkspace}
              placeholder="Provide organization name"
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
              label="Organization URL"
              size="large"
              error={errors.name && String(errors.name?.message)}
              defaultValue={organization?.name || ''}
              disabled
              prefix={URL_PREFIX}
              placeholder="Provide organization URL"
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
