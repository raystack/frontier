import { yupResolver } from '@hookform/resolvers/yup';
import { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import {
  toast,
  Separator,
  Skeleton,
  Box,
  Text,
  Flex,
  Button,
  InputField
} from '@raystack/apsara/v1';
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
          <Flex gap={5} direction="column" style={{ width: '100%' }}>
            <Skeleton
              width="80px"
              height="80px"
              borderRadius={'var(--rs-radius-6)'}
            />
            <Skeleton height="16px" width="100%" />
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
            {isLoading ? (
              <Skeleton height={'32px'} />
            ) : (
              <InputField
                label="Full name"
                size="large"
                error={errors.title && String(errors.title?.message)}
                value={user?.title || ''}
                placeholder="Provide full name"
                {...register('title')}
                disabled={isLoading}
              />
            )}
          </Box>
          <Box style={{ padding: 'var(--pd-4) 0' }}>
            {isLoading ? (
              <Skeleton height={'32px'} />
            ) : (
              <InputField
                label="Email Address"
                size="large"
                error={errors.email && String(errors.email?.message)}
                value={user?.name || ''}
                type="email"
                placeholder="Provide email address"
                {...register('email')}
                readOnly
                disabled
              />
            )}
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
