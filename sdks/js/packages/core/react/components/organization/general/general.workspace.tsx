import { yupResolver } from '@hookform/resolvers/yup';
import {
  Box,
  Button,
  Flex,
  InputField,
  Text,
  TextField,
  Tooltip
} from '@raystack/apsara';
import { forwardRef, useCallback, useEffect, useRef } from 'react';
import { Controller, useForm } from 'react-hook-form';
import Skeleton from 'react-loading-skeleton';
import { toast } from 'sonner';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Organization } from '~/src';
// @ts-ignore
import styles from './general.module.css';
import { AuthTooltipMessage } from '~/react/utils';

const generalSchema = yup
  .object({
    title: yup.string().required('Name is a required field'),
    name: yup.string().required('URL is a required field')
  })
  .required();

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
  const { client } = useFrontier();
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

  async function onSubmit(data: any) {
    if (!client) return;
    if (!organization?.id) return;

    try {
      await client.frontierServiceUpdateOrganization(organization?.id, data);
      toast.success('Updated organization');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
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
            >
              {isSubmitting ? 'updating...' : 'Update'}
            </Button>
          </Tooltip>
        )}
      </Flex>
    </form>
  );
};
