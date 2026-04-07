'use client';

import { useCallback } from 'react';
import { useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { Flex, Button, InputField, toastManager } from '@raystack/apsara-v1';
import { useMutation } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  CreateServiceUserTokenRequestSchema,
  type ServiceUserToken
} from '@raystack/proton/frontier';
import { useFrontier } from '../../../contexts/FrontierContext';
import styles from '../service-account-details-view.module.css';

const tokenSchema = yup
  .object({
    title: yup.string().required('Name is a required field')
  })
  .required();

type FormData = yup.InferType<typeof tokenSchema>;

export interface AddTokenFormProps {
  serviceUserId: string;
  onAddToken: (token: ServiceUserToken) => void;
}

export function AddTokenForm({ serviceUserId, onAddToken }: AddTokenFormProps) {
  const { activeOrganization } = useFrontier();
  const orgId = activeOrganization?.id || '';

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(tokenSchema)
  });

  const { mutateAsync: createServiceUserToken } = useMutation(
    FrontierServiceQueries.createServiceUserToken
  );

  const onSubmit = useCallback(
    async (data: FormData) => {
      try {
        const response = await createServiceUserToken(
          create(CreateServiceUserTokenRequestSchema, {
            orgId,
            id: serviceUserId,
            title: data.title
          })
        );
        if (response.token) {
          onAddToken(response.token);
          reset();
          toastManager.add({ title: 'API key created', type: 'success' });
        }
      } catch (error: unknown) {
        toastManager.add({
          title: 'Something went wrong',
          description: error instanceof Error ? error.message : 'Unknown error',
          type: 'error'
        });
      }
    },
    [createServiceUserToken, onAddToken, serviceUserId, orgId, reset]
  );

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Flex className={styles.addTokenRow} align="start">
        <InputField
          {...register('title')}
          size="large"
          placeholder="Label Name"
          error={errors.title && String(errors.title?.message)}
          className={styles.addTokenInput}
        />
        <Button
          variant="solid"
          color="accent"
          size="normal"
          type="submit"
          loading={isSubmitting}
          disabled={isSubmitting}
          loaderText="Generating..."
          data-test-id="frontier-sdk-service-account-generate-key-btn"
        >
          Generate new key
        </Button>
      </Flex>
    </form>
  );
}
