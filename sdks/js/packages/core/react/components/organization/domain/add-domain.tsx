import {
  Button,
  Image,
  Text,
  Flex,
  toast,
  Dialog,
  InputField
} from '@raystack/apsara';

import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate } from '@tanstack/react-router';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { FrontierServiceQueries, CreateOrganizationDomainRequestSchema, ListOrganizationDomainsRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
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
  const { activeOrganization: organization } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const { mutateAsync: createOrganizationDomain } = useMutation(
    FrontierServiceQueries.createOrganizationDomain,
    {
      onSuccess: async (data) => {
        toast.success('Domain successfully added');
        // Invalidate domains list to refetch
        if (organization?.id) {
          await queryClient.invalidateQueries({
            queryKey: createConnectQueryKey({
              schema: FrontierServiceQueries.listOrganizationDomains,
              transport,
              input: create(ListOrganizationDomainsRequestSchema, {
                orgId: organization.id
              }),
              cardinality: 'finite'
            })
          });
        }
        navigate({
          to: `/domains/$domainId/verify`,
          params: { domainId: data?.domain?.id ?? '' }
        });
      },
      onError: (error: Error) => {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    }
  );

  async function onSubmit(data: FormData) {
    if (!organization?.id) return;

    await createOrganizationDomain(
      create(CreateOrganizationDomainRequestSchema, {
        orgId: organization.id,
        domain: data.domain
      })
    );
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
                Add domain
              </Button>
            </Flex>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
