import { Link, Text, Flex } from '@raystack/apsara/v1';
import React, { ComponentPropsWithRef, useCallback } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { Container } from '../Container';
import { Header } from '../Header';
import { MagicLink } from './magiclink';
import { OIDCButton } from './oidc';

// @ts-ignore
import styles from './onboarding.module.css';

type SignedInProps = ComponentPropsWithRef<typeof Container> & {
  logo?: React.ReactNode;
  title?: string;
  excludes?: string[];
  footer?: boolean;
};
export const SignIn = ({
  logo,
  title = 'Login to Raystack',
  excludes = [],
  footer = true,
  ...props
}: SignedInProps) => {
  const { config } = useFrontier();

  const { data: strategiesData } = useQuery(
    FrontierServiceQueries.listAuthStrategies
  );
  const strategies = strategiesData?.strategies || [];

  const authenticateMutation = useMutation(FrontierServiceQueries.authenticate);

  const clickHandler = useCallback(
    async (name?: string) => {
      if (!name) return;
      try {
        const response = await authenticateMutation.mutateAsync({
          strategyName: name,
          callbackUrl: config.callbackUrl
        });
        if (response.endpoint) {
          window.location.href = response.endpoint;
        }
      } catch (error) {
        console.error('Authentication failed:', error);
      }
    },
    [authenticateMutation, config.callbackUrl]
  );

  const mailotp = strategies.find(s => s.name === 'mailotp');
  const filteredOIDC = strategies
    .filter(s => s.name !== 'mailotp')
    .filter(s => !excludes.includes(s.name ?? ''));

  return (
    <Container {...props}>
      <Header logo={logo} title={title} />
      <Flex direction="column" width="full" gap={5}>
        {filteredOIDC.map((s, index) => {
          return (
            <OIDCButton
              key={index}
              onClick={() => clickHandler(s.name)}
              provider={s.name || ''}
              data-test-id="frontier-sdk-oidc-btn"
            ></OIDCButton>
          );
        })}

        {mailotp && <MagicLink />}
      </Flex>
      {footer && (
        <div style={{ fontWeight: '400' }}>
          <Text size="small">
            Don&apos;t have an account?{' '}
            <Link
              href={config.redirectSignup || ''}
              className={styles.redirectLink}
              data-test-id="frontier-sdk-signup-btn"
            >
              Signup
            </Link>
          </Text>
        </div>
      )}
    </Container>
  );
};
