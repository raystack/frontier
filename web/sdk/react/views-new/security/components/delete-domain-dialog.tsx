'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { create } from '@bufbuild/protobuf';
import { useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import {
  FrontierServiceQueries,
  DeleteOrganizationDomainRequestSchema,
  ListOrganizationDomainsRequestSchema
} from '@raystack/proton/frontier';
import {
  AlertDialog,
  Button,
  Checkbox,
  Flex,
  Text,
  InputField,
  Skeleton,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useOrganizationDomain } from '../../../hooks/useOrganizationDomain';

const domainSchema = yup
  .object({
    domain: yup
      .string()
      .required()
      .matches(/[-a-zA-Z0-9.]{1,256}\.[a-zA-Z0-9()]{1,6}$/, 'Domain is invalid')
  })
  .required();

type FormData = yup.InferType<typeof domainSchema>;

export type DeleteDomainPayload = { domainId: string };

export interface DeleteDomainDialogProps {
  handle: ReturnType<typeof AlertDialog.createHandle<DeleteDomainPayload>>;
  refetch: () => void;
}

export function DeleteDomainDialog({
  handle,
  refetch
}: DeleteDomainDialogProps) {
  const handleOpenChange = (open: boolean) => {
    if (!open) refetch();
  };

  return (
    <AlertDialog handle={handle} onOpenChange={handleOpenChange}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as DeleteDomainPayload | undefined;
        return (
          <DeleteDomainContent
            domainId={payload?.domainId}
            handle={handle}
          />
        );
      }}
    </AlertDialog>
  );
}

function DeleteDomainContent({
  domainId,
  handle
}: {
  domainId: string | undefined;
  handle: DeleteDomainDialogProps['handle'];
}) {
  const { activeOrganization: organization } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();
  const [isAcknowledged, setIsAcknowledged] = useState(false);

  const { domain, isLoading: isDomainLoading } = useOrganizationDomain(domainId);

  const {
    watch,
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors }
  } = useForm({
    resolver: yupResolver(domainSchema)
  });

  const { mutateAsync: deleteOrganizationDomain, isPending } = useMutation(
    FrontierServiceQueries.deleteOrganizationDomain,
    {
      onSuccess: async () => {
        toastManager.add({ title: 'Domain successfully deleted', type: 'success' });
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
        handle.close();
      },
      onError: (error: Error) => {
        toastManager.add({
          title: 'Something went wrong',
          description: error.message,
          type: 'error'
        });
      }
    }
  );

  async function onSubmit(data: FormData) {
    if (!domain?.id || !organization?.id) return;

    if (data.domain !== domain.name) {
      return setError('domain', { message: 'Domain name does not match' });
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
    <AlertDialog.Content width={400}>
      <AlertDialog.Header>
        <AlertDialog.Title>Delete Domain</AlertDialog.Title>
      </AlertDialog.Header>
      <form onSubmit={handleSubmit(onSubmit)}>
        <AlertDialog.Body>
          <Flex direction="column" gap={5}>
            {isDomainLoading ? (
              <>
                <Skeleton height={32} />
                <Skeleton width="50%" height={16} />
                <Skeleton height={32} />
                <Skeleton height={32} />
              </>
            ) : (
              <>
                <Text size="small">
                  This action can not be undone. This will permanently delete{' '}
                  <b>{domain?.name}</b>.
                </Text>
                <InputField
                  label="Please enter the domain name to confirm."
                  size="large"
                  error={errors.domain && String(errors.domain?.message)}
                  {...register('domain')}
                  placeholder="Enter the domain name"
                  data-test-id="frontier-sdk-delete-domain-input"
                />
                <Flex gap={3} align="start">
                  <Checkbox
                    checked={isAcknowledged}
                    onCheckedChange={(v) => setIsAcknowledged(v === true)}
                    data-test-id="frontier-sdk-delete-domain-checkbox"
                  />
                  <Text size="small">
                    I acknowledge and understand that all of the domain data will
                    be deleted and want to proceed.
                  </Text>
                </Flex>
              </>
            )}
          </Flex>
        </AlertDialog.Body>
        <AlertDialog.Footer>
          <Flex justify="end">
            {isDomainLoading ? (
              <Skeleton height={32} width={120} />
            ) : (
              <Button
                variant="solid"
                color="danger"
                type="submit"
                disabled={!domainName || !isAcknowledged}
                loading={isPending}
                loaderText="Deleting..."
                data-test-id="frontier-sdk-delete-domain-btn"
              >
                Delete this domain
              </Button>
            )}
          </Flex>
        </AlertDialog.Footer>
      </form>
    </AlertDialog.Content>
  );
}
