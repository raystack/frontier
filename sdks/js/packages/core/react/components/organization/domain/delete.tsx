import {
  Button,
  Checkbox,
  Skeleton,
  Text,
  Flex,
  toast,
  InputField,
  Dialog
} from '@raystack/apsara';

import { useEffect } from 'react';
import { yupResolver } from '@hookform/resolvers/yup';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import * as yup from 'yup';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { FrontierServiceQueries,
  DeleteOrganizationDomainRequestSchema,
  ListOrganizationDomainsRequestSchema
} from '@raystack/proton/frontier';
import { useOrganizationDomain } from '~/react/hooks/useOrganizationDomain';
import { create } from '@bufbuild/protobuf';
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

export const DeleteDomain = () => {
  const {
    watch,
    handleSubmit,
    setError,
    formState: { errors },
    register
  } = useForm({
    resolver: yupResolver(domainSchema)
  });
  const navigate = useNavigate({ from: '/domains/$domainId/delete' });
  const { domainId } = useParams({ from: '/domains/$domainId/delete' });
  const { activeOrganization: organization } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  const { domain, isLoading, error: domainError } = useOrganizationDomain(domainId);

  useEffect(() => {
    if (domainError) {
      toast.error('Something went wrong', {
        description: (domainError as Error).message
      });
    }
  }, [domainError]);

  const { mutateAsync: deleteOrganizationDomain, isPending } = useMutation(
    FrontierServiceQueries.deleteOrganizationDomain,
    {
      onSuccess: async () => {
        toast.success('Domain successfully deleted');
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
        navigate({ to: '/domains' });
      },
      onError: (error: Error) => {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    }
  );

  async function onSubmit(data: FormData) {
    if (!domain?.id || !organization?.id) return;

    if (data.domain !== domain.name) {
      return setError('domain', { message: 'domain name is not same' });
    }

    await deleteOrganizationDomain(
      create(DeleteOrganizationDomainRequestSchema, {
        id: domain.id,
        orgId: organization.id
      })
    );
  }

  const domainName = watch('domain', '');

  return (
    <Dialog open={true}>
      <Dialog.Content width={600} overlayClassName={styles.overlay}>
        <Dialog.Header>
          <Dialog.Title>Verify domain deletion</Dialog.Title>
          <Dialog.CloseButton
            onClick={() =>
              navigate({
                to: `/domains`
              })
            }
            data-test-id="frontier-sdk-delete-domain-close-btn"
          />
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Body>
            <Flex direction="column" gap={5}>
              {isLoading ? (
                <>
                  <Skeleton height={'16px'} />
                  <Skeleton width={'50%'} height={'16px'} />
                  <Skeleton height={'32px'} />
                  <Skeleton height={'16px'} />
                  <Skeleton height={'32px'} />
                </>
              ) : (
                <>
                  <Text size="small">
                    This action can not be undone. This will permanently delete{' '}
                    <b>{domain?.name}</b>.
                  </Text>

                  <InputField
                    label="Please type the domain name"
                    size="large"
                    error={errors.domain && String(errors.domain?.message)}
                    {...register('domain')}
                    placeholder="Provide domain name"
                  />

                  <Flex gap="small">
                    <Checkbox
                      checked={isAcknowledged}
                      onCheckedChange={v => setIsAcknowledged(v === true)}
                      data-test-id="frontier-sdk-delete-domain-checkbox"
                    />
                    <Text size="small">
                      I acknowledge I understand that all of the team data will
                      be deleted and want to proceed.
                    </Text>
                  </Flex>

                  <Button
                    variant="solid"
                    color="danger"
                    disabled={!domainName || !isAcknowledged || isPending}
                    type="submit"
                    style={{ width: '100%' }}
                    loading={isPending}
                    loaderText="Deleting..."
                    data-test-id="frontier-sdk-delete-domain-btn"
                  >
                    Delete this domain
                  </Button>
                </>
              )}
            </Flex>
          </Dialog.Body>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
