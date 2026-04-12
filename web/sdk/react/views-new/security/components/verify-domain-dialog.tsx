'use client';

import { create } from '@bufbuild/protobuf';
import { useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import {
  FrontierServiceQueries,
  VerifyOrganizationDomainRequestSchema,
  ListOrganizationDomainsRequestSchema
} from '@raystack/proton/frontier';
import {
  Button,
  CopyButton,
  Flex,
  Text,
  Dialog,
  InputField,
  Skeleton,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useOrganizationDomain } from '../../../hooks/useOrganizationDomain';
import { handleConnectError } from '~/utils/error';

export type VerifyDomainPayload = { domainId: string };

export interface VerifyDomainDialogProps {
  handle: ReturnType<typeof Dialog.createHandle<VerifyDomainPayload>>;
  refetch: () => void;
}

export function VerifyDomainDialog({
  handle,
  refetch
}: VerifyDomainDialogProps) {
  const { activeOrganization: organization } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const handleOpenChange = (open: boolean) => {
    if (!open) refetch();
  };

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as VerifyDomainPayload | undefined;
        return (
          <VerifyDomainContent
            domainId={payload?.domainId}
            handle={handle}
            organization={organization}
            queryClient={queryClient}
            transport={transport}
          />
        );
      }}
    </Dialog>
  );
}

function VerifyDomainContent({
  domainId,
  handle,
  organization,
  queryClient,
  transport
}: {
  domainId: string | undefined;
  handle: VerifyDomainDialogProps['handle'];
  organization: ReturnType<typeof useFrontier>['activeOrganization'];
  queryClient: ReturnType<typeof useQueryClient>;
  transport: ReturnType<typeof useTransport>;
}) {
  const { domain, isLoading: isDomainLoading } = useOrganizationDomain(domainId);

  const { mutateAsync: verifyOrganizationDomain, isPending: isVerifying } =
    useMutation(FrontierServiceQueries.verifyOrganizationDomain, {
      onSuccess: async () => {
        toastManager.add({ title: 'Domain verified', type: 'success' });
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
      }
    });

  async function handleVerify() {
    if (!domainId || !organization?.id) return;

    try {
      await verifyOrganizationDomain(
        create(VerifyOrganizationDomainRequestSchema, {
          orgId: organization.id,
          id: domainId
        })
      );
    } catch (error) {
      handleConnectError(error, {
        NotFound: (err) => toastManager.add({ title: 'Not found', description: err.message, type: 'error' }),
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
        InvalidArgument: (err) => toastManager.add({ title: 'Invalid input', description: err.message, type: 'error' }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
      });
    }
  }

  return (
    <Dialog.Content width={400}>
      <Dialog.Header>
        <Dialog.Title>Verify domain</Dialog.Title>
      </Dialog.Header>
      <Dialog.Body>
        <Flex direction="column" gap={5}>
          {isDomainLoading ? (
            <>
              <Skeleton height={32} />
              <Skeleton height={36} />
              <Skeleton height={32} />
            </>
          ) : (
            <>
              <Text size="small">
                Before we can verify {domain?.name}, you&apos;ll need to create
                a TXT record in your DNS configuration for this hostname.
              </Text>
              <InputField
                value={domain?.token || ''}
                size="large"
                readOnly
                trailingIcon={
                  domain?.token ? (
                    <CopyButton
                      text={domain.token}
                      size={2}
                      data-test-id="frontier-sdk-domain-token-copy-btn"
                    />
                  ) : undefined
                }
                data-test-id="frontier-sdk-domain-token-input"
              />
              <Text size="small" variant="secondary">
                Wait until your DNS configuration changes. This could take up to
                72 hours to propagate.
              </Text>
            </>
          )}
        </Flex>
      </Dialog.Body>
      <Dialog.Footer>
        <Flex justify="end">
          {isDomainLoading ? (
            <Skeleton height={32} width={64} />
          ) : (
            <Button
              variant="solid"
              color="accent"
              onClick={handleVerify}
              loading={isVerifying}
              loaderText="Verifying..."
              data-test-id="frontier-sdk-verify-domain-btn"
            >
              Verify
            </Button>
          )}
        </Flex>
      </Dialog.Footer>
    </Dialog.Content>
  );
}
