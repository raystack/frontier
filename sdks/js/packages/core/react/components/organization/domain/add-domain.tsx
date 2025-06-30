import {
  Button,
  Image,
  Text,
  Flex,
  toast,
  Dialog,
  InputField
} from '@raystack/apsara/v1';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { useForm } from 'react-hook-form';
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
    register,
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
      <Dialog.Content
        overlayClassName={styles.overlay}
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
      >
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Header>
            <Flex justify="between" align="center" style={{ width: '100%' }}>
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
          </Dialog.Header>

          <Dialog.Body>
            <Flex direction="column" gap={5}>
              <InputField
                label="Domain name"
                size="large"
                error={errors.domain && String(errors.domain?.message)}
                {...register('domain')}
                name="domain"
                placeholder="Provide domain name"
              />
            </Flex>
          </Dialog.Body>

          <Dialog.Footer>
            <Flex justify="end">
              <Button
                type="submit"
                loading={isSubmitting}
                loaderText="Adding..."
                data-test-id="frontier-sdk-add-domain-btn"
              >
                Add Domain
              </Button>
            </Flex>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
