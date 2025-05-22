import { Dialog, Flex, Separator, Text } from '@raystack/apsara';
import { Button, Skeleton, Image } from '@raystack/apsara/v1';

import { useCallback, useEffect, useState } from 'react';

import { useNavigate, useParams } from '@tanstack/react-router';
import { toast } from '@raystack/apsara/v1';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Domain } from '~/src';
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
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} style={{ fontWeight: '500' }}>
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

        <Flex direction="column" gap="medium" style={{ padding: '24px 32px' }}>
          {isDomainLoading ? (
            <>
              <Skeleton height={'16px'} />
              <Skeleton height={'32px'} />
              <Skeleton height={'16px'} />
            </>
          ) : (
            <>
              <Text size={2}>
                Before we can verify {domain?.name}, you&apos;ll need to create
                a TXT record in your DNS configuration for this hostname.
              </Text>
              <Flex
                style={{
                  padding: 'var(--pd-8)',
                  border: '1px solid var(--border-base)',
                  borderRadius: 'var(--pd-4)'
                }}
              >
                <Text size={2}>{domain?.token}</Text>
              </Flex>
              <Text size={2}>
                Wait until your DNS configuration changes. This could take up to
                72 hours to propagate.
              </Text>
            </>
          )}
        </Flex>
        <Separator />
        <Flex justify="end" style={{ padding: 'var(--pd-16)' }}>
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
      </Dialog.Content>
    </Dialog>
  );
};
