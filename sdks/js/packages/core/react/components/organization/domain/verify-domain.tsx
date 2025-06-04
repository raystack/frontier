import { useCallback, useEffect, useState } from 'react';
import { Button, Separator, Skeleton, Image, Text, Flex, toast, Dialog } from '@raystack/apsara/v1';

import { useNavigate, useParams } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Domain } from '~/src';
import cross from '~/react/assets/cross.svg';
import styles from '../organization.module.css';

export const VerifyDomain = () => {
  const navigate = useNavigate({ from: '/domains/$domainId/verify' });
  const { domainId } = useParams({ from: '/domains/$domainId/verify' });
  const { client, activeOrganization: organization } = useFrontier();
  const [domain, setDomain] = useState<V1Beta1Domain>();
  const [isVerifying, setIsVerifying] = useState(false);
  const [isDomainLoading, setIsDomainLoading] = useState(false);

  const fetchDomainDetails = useCallback(async () => {
    if (!domainId) return;
    if (!organization?.id) return;

    try {
      setIsDomainLoading(true);
      const resp = await client?.frontierServiceGetOrganizationDomain(organization?.id, domainId);
      const domain = resp?.data.domain
      setDomain(domain);
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    } finally {
      setIsDomainLoading(false);
    }
  }, [client, domainId, organization?.id]);

  const verifyDomain = useCallback(async () => {
    if (!domainId) return;
    if (!organization?.id) return;
    setIsVerifying(true);

    try {
      await client?.frontierServiceVerifyOrganizationDomain(
        organization?.id,
        domainId,
        {}
      );
      navigate({ to: '/domains' });
      toast.success('Domain verified');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    } finally {
      setIsVerifying(false);
    }
  }, [client, domainId, navigate, organization?.id]);

  useEffect(() => {
    fetchDomainDetails();
  }, [fetchDomainDetails]);

  return (
    <Dialog open={true}>
      <Dialog.Content overlayClassName={styles.overlay} style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}>
        <Dialog.Header>
          <Flex justify="between" style={{ padding: '16px 24px' }}>
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
          <Separator />
        </Dialog.Header>

        <Dialog.Body>
          <Flex direction="column" gap={5} style={{ padding: 'var(--rs-space-7) var(--rs-space-9)' }}>
            {isDomainLoading ? (
              <>
                <Skeleton height={'16px'} />
                <Skeleton height={'32px'} />
                <Skeleton height={'16px'} />
              </>
            ) : (
              <>
                <Text size="small">
                  Before we can verify {domain?.name}, you&apos;ll need to create
                  a TXT record in your DNS configuration for this hostname.
                </Text>
                <Flex
                  style={{
                    padding: 'var(--rs-space-3)',
                    border: '1px solid var(--rs-color-border-base-secondary)',
                    borderRadius: 'var(--rs-space-2)'
                  }}
                >
                  <Text size="small">{domain?.token}</Text>
                </Flex>
                <Text size="small">
                  Wait until your DNS configuration changes. This could take up to
                  72 hours to propagate.
                </Text>
              </>
            )}
          </Flex>
          <Separator />
        </Dialog.Body>

        <Dialog.Footer>
          <Flex justify="end" style={{ padding: 'var(--rs-space-5)' }}>
            {isDomainLoading ? (
              <Skeleton height={'32px'} width={'64px'} />
            ) : (
              <Button
                onClick={verifyDomain}
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
