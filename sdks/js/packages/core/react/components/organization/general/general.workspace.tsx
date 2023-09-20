import { yupResolver } from '@hookform/resolvers/yup';
import {
  Box,
  Button,
  Flex,
  InputField,
  Text,
  TextField
} from '@raystack/apsara';
import { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import Skeleton from 'react-loading-skeleton';
import { toast } from 'sonner';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Organization } from '~/src';

const generalSchema = yup
  .object({
    title: yup.string().required(),
    name: yup.string().required()
  })
  .required();

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
          <InputField label="Organization title">
            {isLoading ? (
              <Skeleton height={'32px'} />
            ) : (
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide organization title"
                  />
                )}
                defaultValue={organization?.title}
                control={control}
                name="title"
              />
            )}

            <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
              {errors.title && String(errors.title?.message)}
            </Text>
          </InputField>
        </Box>
        <Box style={{ padding: 'var(--pd-4) 0' }}>
          <InputField label="Organization name">
            {isLoading ? (
              <Skeleton height={'32px'} />
            ) : (
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide organization name"
                    disabled
                  />
                )}
                defaultValue={organization?.name}
                control={control}
                name="name"
              />
            )}

            <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
              {errors.name && String(errors.name?.message)}
            </Text>
          </InputField>
        </Box>
        {canUpdateWorkspace ? (
          <Button
            size="medium"
            variant="primary"
            type="submit"
            style={{ width: 'fit-content' }}
            disabled={isLoading || isSubmitting}
          >
            {isSubmitting ? 'updating...' : 'Update'}
          </Button>
        ) : null}
      </Flex>
    </form>
  );
};
