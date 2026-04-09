'use client';

import { useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { create } from '@bufbuild/protobuf';
import { useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import {
  FrontierServiceQueries,
  CreateOrganizationDomainRequestSchema,
  ListOrganizationDomainsRequestSchema
} from '@raystack/proton/frontier';
import {
  Button,
  Flex,
  Text,
  Dialog,
  InputField,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { handleConnectError } from '~/utils/error';

const domainSchema = yup
  .object({
    domain: yup
      .string()
      .required('Domain name is required')
      .matches(
        /[-a-zA-Z0-9.]{1,256}\.[a-zA-Z0-9()]{1,6}$/,
        'Domain is invalid'
      )
  })
  .required();

type FormData = yup.InferType<typeof domainSchema>;

export interface AddDomainDialogProps {
  handle: ReturnType<typeof Dialog.createHandle>;
  onDomainAdded: (domainId: string) => void;
}

export function AddDomainDialog({
  handle,
  onDomainAdded
}: AddDomainDialogProps) {
  const { activeOrganization: organization } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting, isDirty }
  } = useForm({
    resolver: yupResolver(domainSchema)
  });

  const { mutateAsync: createOrganizationDomain } = useMutation(
    FrontierServiceQueries.createOrganizationDomain,
    {
      onSuccess: async (data) => {
        toastManager.add({ title: 'Domain successfully added', type: 'success' });
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
        onDomainAdded(data?.domain?.id ?? '');
      }
    }
  );

  async function onSubmit(data: FormData) {
    if (!organization?.id) return;

    try {
      await createOrganizationDomain(
        create(CreateOrganizationDomainRequestSchema, {
          orgId: organization.id,
          domain: data.domain
        })
      );
    } catch (error) {
      handleConnectError(error, {
        AlreadyExists: () => toastManager.add({ title: 'Domain already exists', type: 'error' }),
        InvalidArgument: (err) => toastManager.add({ title: 'Invalid input', description: err.message, type: 'error' }),
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
      });
    }
  }

  const handleOpenChange = (open: boolean) => {
    if (!open) reset();
  };

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      <Dialog.Content width={400}>
          <Dialog.Header>
            <Dialog.Title>Add domain</Dialog.Title>
          </Dialog.Header>
          <form onSubmit={handleSubmit(onSubmit)}>
            <Dialog.Body>
              <Flex direction="column" gap={5}>
                <Text size="small">
                  Adding a domain to allowed domains will result in charges for each user onboarded under that domain.
                </Text>
                <InputField
                  label="Domain name"
                  size="large"
                  error={errors.domain && String(errors.domain?.message)}
                  {...register('domain')}
                  placeholder="example.com"
                  data-test-id="frontier-sdk-add-domain-input"
                />
              </Flex>
            </Dialog.Body>
            <Dialog.Footer>
              <Flex justify="end">
                <Button
                  type="submit"
                  variant="solid"
                  color="accent"
                  disabled={!isDirty}
                  loading={isSubmitting}
                  loaderText="Adding..."
                  data-test-id="frontier-sdk-add-domain-submit-btn"
                >
                  Continue
                </Button>
              </Flex>
            </Dialog.Footer>
          </form>
        </Dialog.Content>
    </Dialog>
  );
}
