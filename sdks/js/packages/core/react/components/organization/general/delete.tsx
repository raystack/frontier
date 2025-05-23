import {
  Dialog,
  Flex,
  InputField,
  Text,
  TextField
} from '@raystack/apsara';
import { Button, Checkbox, Separator, toast, Image } from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';

// @ts-ignore
import styles from './general.module.css';
import { useState } from 'react';

const orgSchema = yup
  .object({
    name: yup.string()
  })
  .required();

export const DeleteOrganization = () => {
  const {
    watch,
    control,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(orgSchema)
  });
  const navigate = useNavigate({ from: '/delete' });
  const { client, activeOrganization: organization } = useFrontier();
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  async function onSubmit(data: any) {
    if (!client) return;
    if (!organization?.id) return;
    if (data.name !== organization.name)
      return setError('name', { message: 'organization name is not same' });

    try {
      await client.frontierServiceDeleteOrganization(organization?.id);
      toast.success('Organization deleted');

      // @ts-ignore
      window.location = window.location.origin;
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  const name = watch('name', '');
  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} style={{ fontWeight: '500' }}>
            Verify organization deletion
          </Text>

          <Image
            className={styles.deleteIcon}
            alt="cross"
            style={{ cursor: 'pointer' }}
            src={cross as unknown as string}
            onClick={() => navigate({ to: '/' })}
            data-test-id="frontier-sdk-delete-organization-close-btn"
          />
        </Flex>
        <Separator />
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex
            direction="column"
            gap="medium"
            style={{ padding: '24px 32px' }}
          >
            <Text size={2}>
              This action <b>can not</b> be undone. This will permanently delete
              all the projects and resources in <b>{organization?.title}</b>.
            </Text>

            <InputField label="Please type name of the organization to confirm.">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide organization name"
                  />
                )}
                control={control}
                name="name"
              />

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.name && String(errors.name?.message)}
              </Text>
            </InputField>
            <Flex gap="small">
              <Checkbox
                checked={isAcknowledged}
                onCheckedChange={v => setIsAcknowledged(v === true)}
                data-test-id="frontier-sdk-delete-organization-checkbox" />
              <Text size={2}>
                I acknowledge I understand that all of the organization data
                will be deleted and want to proceed.
              </Text>
            </Flex>

            <Button
              variant="solid"
              color="danger"
              type="submit"
              disabled={!name || !isAcknowledged}
              style={{ width: '100%' }}
              data-test-id="frontier-sdk-delete-organization-btn"
              loading={isSubmitting}
              loaderText="Deleting..."
            >
              Delete this organization
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
