import {
  Dialog,
  InputField,
  TextField
} from '@raystack/apsara';
import { Button, Separator, Image, Text, Flex, toast } from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { Controller, useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import cross from '~/react/assets/cross.svg';
import styles from '../organization.module.css';

const domainSchema = yup
  .object({
    domain: yup
      .string()
      .required()
      .matches(/[-a-zA-Z0-9.]{1,256}\.[a-zA-Z0-9()]{1,6}$/, 'Domain is invalid')
  })
  .required();

type FormData = yup.InferType<typeof domainSchema>;

export const AddDomain = () => {
  const {
    control,
    handleSubmit,
    formState: { errors, isSubmitting }
  } = useForm({
    resolver: yupResolver(domainSchema)
  });
  const navigate = useNavigate({ from: '/domains/modal' });
  const { client, activeOrganization: organization } = useFrontier();

  async function onSubmit(data: FormData) {
    if (!client) return;
    if (!organization?.id) return;

    try {
      const {
        data: { domain }
      } = await client.frontierServiceCreateOrganizationDomain(
        organization?.id,
        data
      );
      toast.success('Domain added');

      navigate({ to: '/domains' });
      navigate({
        to: `/domains/$domainId/verify`,
        params: { domainId: domain?.id ?? '' }
      });
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <form onSubmit={handleSubmit(onSubmit)}>
          <Flex justify="between" style={{ padding: '16px 24px' }}>
            <Text size="large" weight="medium">
              Add domain
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={() => navigate({ to: '/domains' })}
              data-test-id="frontier-sdk-add-domain-btn"
            />
          </Flex>
          <Separator />

          <Flex
            direction="column"
            gap="medium"
            style={{ padding: '24px 32px' }}
          >
            <InputField label="Domain name">
              <Controller
                render={({ field }) => (
                  <TextField
                    {...field}
                    // @ts-ignore
                    size="medium"
                    placeholder="Provide domain name"
                  />
                )}
                control={control}
                name="domain"
              />

              <Text size="mini" variant="danger">
                {errors.domain && String(errors.domain?.message)}
              </Text>
            </InputField>
          </Flex>
          <Separator />
          <Flex justify="end" style={{ padding: 'var(--rs-space-5)' }}>
            <Button
              type="submit"
              loading={isSubmitting}
              loaderText="Adding..."
              data-test-id="frontier-sdk-add-domain-btn"
            >
              Add domain
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
