import {
  Dialog,
  Separator,
  Image,
  InputField,
  TextField
} from '@raystack/apsara';
import styles from './styles.module.css';
import { Button, Flex, Text, toast } from '@raystack/apsara/v1';
import cross from '~/react/assets/cross.svg';
import { useNavigate } from '@tanstack/react-router';
import { Controller, useForm } from 'react-hook-form';
import { useFrontier } from '~/react/contexts/FrontierContext';
import * as yup from 'yup';
import { yupResolver } from '@hookform/resolvers/yup';
import { useCallback } from 'react';

const serviceAccountSchema = yup
  .object({
    title: yup.string().required('Name is a required field')
  })
  .required();

type FormData = yup.InferType<typeof serviceAccountSchema>;

export const AddServiceAccount = () => {
  const navigate = useNavigate({ from: '/api-keys/add' });
  const { client, activeOrganization: organization } = useFrontier();

  const {
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(serviceAccountSchema)
  });

  const orgId = organization?.id;

  const onSubmit = useCallback(
    async (data: FormData) => {
      if (!client) return;
      if (!orgId) return;

      try {
        const {
          data: { serviceuser }
        } = await client.frontierServiceCreateServiceUser({
          body: data,
          org_id: orgId
        });
        toast.success('Service user created');

        navigate({
          to: '/api-keys/$id',
          params: { id: serviceuser?.id ?? '' }
        });
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    },
    [client, navigate, orgId]
  );

  const isDisabled = isSubmitting;

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        overlayClassname={styles.overlay}
        className={styles.addDialogContent}
      >
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex justify="between" className={styles.addDialogForm}>
            <Text size={6} weight={500}>
              New Service Account
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              // @ts-ignore
              src={cross}
              onClick={() => navigate({ to: '/api-keys' })}
              data-test-id="frontier-sdk-new-service-account-close-btn"
            />
          </Flex>
          <Separator />

          <Flex
            direction="column"
            gap="medium"
            className={styles.addDialogFormContent}
          >
            <Text>
              Create a dedicated service account to facilitate secure API
              interactions on behalf of the organization.
            </Text>

            <InputField label="Name">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    size="medium"
                    placeholder="Provide service account name"
                  />
                )}
                name="title"
                control={control}
              />
              <Text size={1} variant="danger">
                {errors.title && String(errors.title?.message)}
              </Text>
            </InputField>
          </Flex>
          <Separator />
          <Flex justify="end" className={styles.addDialogFormBtnWrapper}>
            <Button
              variant="primary"
              size="normal"
              type="submit"
              data-test-id="frontier-sdk-add-service-account-btn"
              loading={isSubmitting}
              disabled={isDisabled}
              loaderText={'Creating...'}
            >
              Create
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
