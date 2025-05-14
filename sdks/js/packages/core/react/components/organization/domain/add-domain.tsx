import {
  Button,
  Dialog,
  Flex,
  Image,
  InputField,
  Separator,
  Text,
  TextField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { Controller, useForm } from 'react-hook-form';
import { toast } from '@raystack/apsara/v1';
import * as yup from 'yup';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
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
            <Text size={6} style={{ fontWeight: '500' }}>
              Add domain
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              // @ts-ignore
              src={cross}
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

              <Text size={1} style={{ color: 'var(--foreground-danger)' }}>
                {errors.domain && String(errors.domain?.message)}
              </Text>
            </InputField>
          </Flex>
          <Separator />
          <Flex justify="end" style={{ padding: 'var(--pd-16)' }}>
            <Button
              variant="primary"
              size="medium"
              type="submit"
              data-test-id="frontier-sdk-add-domain-btn"
            >
              {isSubmitting ? 'Adding...' : 'Add domain'}
            </Button>
          </Flex>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
