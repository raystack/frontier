'use client';

import { useEffect } from 'react';
import { yupResolver } from '@hookform/resolvers/yup';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import {
  createConnectQueryKey,
  useMutation
} from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { useQueryClient } from '@tanstack/react-query';
import {
  Button,
  Flex,
  InputField,
  Skeleton,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../contexts/FrontierContext';
import { ViewContainer } from '../../components/view-container';
import { ViewHeader } from '../../components/view-header';
import { ImageUpload } from '../../components/image-upload';
import styles from './profile-view.module.css';

const profileSchema = yup
  .object({
    avatar: yup.string().optional(),
    title: yup
      .string()
      .required('Name is required')
      .min(2, 'Name must be at least 2 characters')
      .matches(
        /^[\p{L} .'-]+$/u,
        'Name can only contain letters, spaces, periods, hyphens, and apostrophes'
      )
      .matches(/^\p{L}/u, 'Name must start with a letter')
      .matches(
        /^\p{L}[\p{L} .'-]*\p{L}$|^\p{L}$/u,
        'Name must end with a letter'
      )
      .matches(/^(?!.* {2}).*$/, 'Name cannot have consecutive spaces')
      .matches(/^(?!.* [^\p{L}]).*$/u, 'Spaces must be followed by a letter')
      .matches(/^(?!.*-[^\p{L}]).*$/u, 'Hyphens must be followed by a letter')
      .matches(
        /^(?!.*'[^\p{L}]).*$/u,
        'Apostrophes must be followed by a letter'
      ),
    email: yup.string().email().required()
  })
  .required();

type FormData = yup.InferType<typeof profileSchema>;

export function ProfileView() {
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
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isSubmitting, isDirty }
  } = useForm({
    resolver: yupResolver(profileSchema)
  });

  useEffect(() => {
    reset(user, { keepDirtyValues: true });
  }, [user, reset]);

  async function onSubmit(data: FormData) {
    try {
      if (!user?.id) return;
      await updateCurrentUser({ body: data });
      toastManager.add({ title: 'Updated user', type: 'success' });
    } catch (err: unknown) {
      toastManager.add({
        title: 'Something went wrong',
        description: err instanceof Error ? err.message : 'Failed to update',
        type: 'error'
      });
    }
  }

  return (
    <ViewContainer>
      <ViewHeader
        title="Profile"
        description="Manage your profile information and settings."
      />

      <form onSubmit={handleSubmit(onSubmit)}>
        <Flex direction="column">
          {/* Avatar section */}
          <Flex direction="column" gap={5} className={styles.section}>
            {isLoading ? (
              <Flex direction="column" gap={5}>
                <Skeleton width="72px" height="72px" />
                <Skeleton height="20px" width="50%" />
              </Flex>
            ) : (
              <ImageUpload
                value={watch('avatar')}
                onChange={(value: string) =>
                  setValue('avatar', value, { shouldDirty: true })
                }
                description="Pick a profile picture for your avatar. Max size: 5 Mb"
                initials={user?.title?.[0]}
                data-test-id="frontier-sdk-profile-avatar-upload"
              />
            )}
          </Flex>

          {/* Form section */}
          <Flex direction="column" gap={7} className={styles.section}>
            <Flex direction="column" gap={9} className={styles.formFields}>
              {isLoading ? (
                <>
                  <Skeleton height="58px" />
                  <Skeleton height="58px" />
                </>
              ) : (
                <>
                  <InputField
                    label="Full name"
                    size="large"
                    error={errors.title && String(errors.title?.message)}
                    defaultValue={user?.title || ''}
                    placeholder="Provide full name"
                    {...register('title')}
                    disabled={isLoading}
                  />
                  <InputField
                    label="Email address"
                    size="large"
                    error={errors.email && String(errors.email?.message)}
                    value={user?.email || ''}
                    type="email"
                    placeholder="Provide email address"
                    {...register('email')}
                    readOnly
                    disabled
                  />
                </>
              )}
            </Flex>

            {isLoading ? (
              <Skeleton height="32px" width="64px" />
            ) : (
              <Button
                type="submit"
                variant="solid"
                color="accent"
                disabled={isLoading || isSubmitting || !isDirty}
                loading={isSubmitting}
                loaderText="Updating..."
                data-test-id="frontier-sdk-update-user-btn"
              >
                Update
              </Button>
            )}
          </Flex>
        </Flex>
      </form>
    </ViewContainer>
  );
}
