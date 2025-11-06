import { yupResolver } from '@hookform/resolvers/yup';
import { useEffect } from 'react';
import { Controller, useForm } from 'react-hook-form';
import {
  toast,
  Separator,
  Skeleton,
  Box,
  Flex,
  Button,
  InputField
} from '@raystack/apsara';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { AvatarUpload } from '../../avatar-upload';
import { useMutation, FrontierServiceQueries } from '~hooks';
import { useQueryClient } from '@tanstack/react-query';
import { createConnectQueryKey } from '@connectrpc/connect-query';

const generalSchema = yup
  .object({
    avatar: yup.string().optional(),
    title: yup.string().required('Name is required'),
    email: yup.string().email().required()
  })
  .required();

type FormData = yup.InferType<typeof generalSchema>;

export const UpdateProfile = () => {
  const { user, isUserLoading: isLoading } = useFrontier();
  const queryClient = useQueryClient();
  const { mutateAsync: updateCurrentUser } = useMutation(
    FrontierServiceQueries.updateCurrentUser,
    {
      onSuccess: () => {
        queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.getCurrentUser,
            cardinality: 'finite'
          })
        });
      }
    }
  );
  const {
    reset,
    control,
    register,
    handleSubmit,
    formState: { errors, isSubmitting, isDirty }
  } = useForm({
    resolver: yupResolver(generalSchema)
  });

  useEffect(() => {
    reset(user, { keepDirtyValues: true });
  }, [user, reset]);

  async function onSubmit(data: FormData) {
    try {
      if (!user?.id) return;

      await updateCurrentUser({
        body: data
      });
      toast.success('Updated user');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Flex direction="column" gap={9}>
        <Flex>
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
        <Flex direction="column" gap="large">
          <Flex direction="column" gap="large" style={{ maxWidth: '320px' }}>
            <Box style={{ padding: 'var(--pd-4) 0' }}>
              {isLoading ? (
                <Skeleton height={'32px'} />
              ) : (
                <InputField
                  label="Full name"
                  size="large"
                  error={errors.title && String(errors.title?.message)}
                  defaultValue={user?.title || ''}
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
                  value={user?.email || ''}
                  type="email"
                  placeholder="Provide email address"
                  {...register('email')}
                  readOnly
                  disabled
                />
              )}
            </Box>
            <Button
              size="normal"
              type="submit"
              style={{ width: 'fit-content' }}
              disabled={isLoading || isSubmitting || !isDirty}
              loaderText="Updating..."
              data-test-id="frontier-sdk-update-user-btn"
            >
              Update
            </Button>
          </Flex>
        </Flex>
      </Flex>
    </form>
  );
};
