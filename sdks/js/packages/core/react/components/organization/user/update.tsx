import { yupResolver } from '@hookform/resolvers/yup';
import {
  Box,
  Button,
  Flex,
  InputField,
  Separator,
  Text,
  TextField
} from '@raystack/apsara';
import { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import Skeleton from 'react-loading-skeleton';
import { toast } from 'sonner';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { AvatarUpload } from '../../avatar-upload';
import { styles } from '../styles';

const generalSchema = yup
  .object({
    avatar: yup.string().optional(),
    title: yup.string().required('Name is required'),
    email: yup.string().email().required()
  })
  .required();

type FormData = yup.InferType<typeof generalSchema>;

export const UpdateProfile = () => {
  const { client, user, isUserLoading: isLoading, setUser } = useFrontier();
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

  async function onSubmit(data: FormData) {
    try {
      if (!client) return;
      if (!user?.id) return;

      // This API call can be moved to sdks/js/packages/core/react/components/organization/user/index.tsx
      const updatedUser = await client.frontierServiceUpdateCurrentUser(data);
      if (updatedUser?.data?.user) {
        setUser(updatedUser?.data?.user);
      }
      toast.success('Updated user');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Flex style={styles.container}>
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
                subText="Pick a profile picture for your avatar"
              />
            )}
            control={control}
            name="avatar"
          />
        )}
      </Flex>
      <Separator />
      <Flex direction="column" gap="large" style={styles.container}>
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
            data-test-id="frontier-sdk-update-user-btn"
          >
            {isSubmitting ? 'Updating...' : 'Update'}
          </Button>
        </Flex>
      </Flex>
    </form>
  );
};
