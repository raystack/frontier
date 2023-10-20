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

const generalSchema = yup
  .object({
    title: yup.string().required('Name is required'),
    email: yup.string().email().required()
  })
  .required();

export const UpdateProfile = () => {
  const { client, user, isUserLoading: isLoading } = useFrontier();
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
    reset(user);
  }, [user, reset]);

  async function onSubmit(data: any) {
    try {
      if (!client) return;
      if (!user?.id) return;

      await client.frontierServiceUpdateCurrentUser(data);
      toast.success('Updated user');
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
          <InputField label="Full name">
            {isLoading ? (
              <Skeleton height={'32px'} />
            ) : (
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide full name"
                  />
                )}
                defaultValue={user?.title}
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
          <InputField label="Email Address">
            {isLoading ? (
              <Skeleton height={'32px'} />
            ) : (
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    type="email"
                    // @ts-ignore
                    size="medium"
                    readOnly
                    disabled
                    placeholder="Provide email address"
                  />
                )}
                defaultValue={user?.name}
                control={control}
                name="email"
              />
            )}

            <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
              {errors.email && String(errors.email?.message)}
            </Text>
          </InputField>
        </Box>
        <Button
          size="medium"
          variant="primary"
          type="submit"
          style={{ width: 'fit-content' }}
          disabled={isLoading || isSubmitting}
        >
          {isSubmitting ? 'updating...' : 'Update'}
        </Button>
      </Flex>
    </form>
  );
};
