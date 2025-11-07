import {
  Button,
  Skeleton,
  Image,
  Text,
  Flex,
  toast,
  Dialog,
  CopyButton
} from '@raystack/apsara';

import { useEffect } from 'react';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useQueryClient } from '@tanstack/react-query';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { FrontierServiceQueries,
  VerifyOrganizationDomainRequestSchema,
  ListOrganizationDomainsRequestSchema
} from '@raystack/proton/frontier';
import { useOrganizationDomain } from '~/react/hooks/useOrganizationDomain';
import { create } from '@bufbuild/protobuf';
import cross from '~/react/assets/cross.svg';
import styles from '../organization.module.css';

export const VerifyDomain = () => {
  const navigate = useNavigate({ from: '/domains/$domainId/verify' });
  const { domainId } = useParams({ from: '/domains/$domainId/verify' });
  const { activeOrganization: organization } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const { domain, isLoading: isDomainLoading, error: domainError } = useOrganizationDomain(domainId);

  useEffect(() => {
    if (domainError) {
      toast.error('Something went wrong', {
        description: (domainError as Error).message
      });
    }
  }, [domainError]);

  const { mutateAsync: verifyOrganizationDomain, isPending: isVerifying } = useMutation(
    FrontierServiceQueries.verifyOrganizationDomain,
    {
      onSuccess: async () => {
        toast.success('Domain verified');
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

  async function handleVerify() {
    if (!domainId || !organization?.id) return;

    await verifyOrganizationDomain(
      create(VerifyOrganizationDomainRequestSchema, {
        orgId: organization.id,
        id: domainId
      })
    );
  }

  return (
    <Dialog open={true}>
      <Dialog.Content
        overlayClassName={styles.overlay}
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Verify domain
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={() => navigate({ to: '/domains' })}
              data-test-id="frontier-sdk-verify-domain-close-btn"
            />
          </Flex>
        </Dialog.Header>

        <Dialog.Body>
          <Flex direction="column" gap={5}>
            {isDomainLoading ? (
              <>
                <Skeleton height={'16px'} />
                <Skeleton height={'32px'} />
                <Skeleton height={'16px'} />
              </>
            ) : (
              <>
                <Text size="small">
                  Before we can verify {domain?.name}, you&apos;ll need to
                  create a TXT record in your DNS configuration for this
                  hostname.
                </Text>
                <Flex
                  justify="between"
                  align="center"
                  gap={3}
                  style={{
                    padding: 'var(--rs-space-3)',
                    border: '1px solid var(--rs-color-border-base-secondary)',
                    borderRadius: 'var(--rs-space-2)'
                  }}
                >
                  <Text size="small" style={{ wordBreak: 'break-all' }}>
                    {domain?.token}
                  </Text>
                  {domain?.token && (
                    <CopyButton
                      text={domain.token}
                      size={3}
                      data-test-id="frontier-sdk-domain-token-copy-btn"
                    />
                  )}
                </Flex>
                <Text size="small">
                  Wait until your DNS configuration changes. This could take up
                  to 72 hours to propagate.
                </Text>
              </>
            )}
          </Flex>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex justify="end">
            {isDomainLoading ? (
              <Skeleton height={'32px'} width={'64px'} />
            ) : (
              <Button
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
    </Dialog>
  );
};
