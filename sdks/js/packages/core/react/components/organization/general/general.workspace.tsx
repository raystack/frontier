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
  organization
}: {
  organization?: V1Beta1Organization;
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
        <TextField
          name="workspaceID"
          defaultValue={organization?.id}
          hidden={true}
        />
        <Box style={{ padding: 'var(--pd-4) 0' }}>
          <InputField label="organization title">
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

            <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
              {errors.title && String(errors.title?.message)}
            </Text>
          </InputField>
        </Box>
        <Box style={{ padding: 'var(--pd-4) 0' }}>
          <InputField label="organization name">
            <Controller
              render={({ field }) => (
                <TextField
                  {...field}
                  // @ts-ignore
                  size="medium"
                  placeholder="Provide organization name"
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
        </Box>
        <Button
          size="medium"
          variant="primary"
          type="submit"
          style={{ width: 'fit-content' }}
        >
          {isSubmitting ? 'updating...' : 'Update'}
        </Button>
      </Flex>
    </form>
  );
};
