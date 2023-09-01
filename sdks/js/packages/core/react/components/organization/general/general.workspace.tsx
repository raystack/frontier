import { yupResolver } from '@hookform/resolvers/yup';
import {
  Box,
  Button,
  Flex,
  InputField,
  Text,
  TextField
} from '@raystack/apsara';
import { useCallback, useEffect } from 'react';
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
  const { client, setOrganizations } = useFrontier();
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

  const updateOrganizations = useCallback(
    (updatedOrg: V1Beta1Organization) => {
      setOrganizations(prev =>
        prev.map(org => (org.id === updatedOrg.id ? updatedOrg : org))
      );
    },
    [setOrganizations]
  );

  async function onSubmit(data: any) {
    if (!client) return;
    if (!organization?.id) return;

    try {
      const {
        data: { organization: updatedOrganization }
      } = await client.frontierServiceUpdateOrganization(
        organization?.id,
        data
      );
      if (updatedOrganization) {
        toast.success('Updated organization');
        updateOrganizations(updatedOrganization);
      }
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
