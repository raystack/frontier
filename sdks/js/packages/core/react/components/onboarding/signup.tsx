import { Link, Text, Flex } from '@raystack/apsara/v1';
import React, { ComponentPropsWithRef, useCallback } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { Container } from '../Container';
import { Header } from '../Header';
import { MagicLink } from './magiclink';
import { OIDCButton } from './oidc';

// @ts-ignore
import styles from './onboarding.module.css';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { useQuery, useMutation } from '@connectrpc/connect-query';

type SignUpProps = ComponentPropsWithRef<typeof Container> & {
  logo?: React.ReactNode;
  title?: string;
  excludes?: string[];
};
export const SignUp = ({
  logo,
  title = 'Create your account',
  excludes = [],
  ...props
}: SignUpProps) => {
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
      <Flex direction="column" gap={3} width="full">
        {filteredOIDC.map((s, index) => {
          return (
            <OIDCButton
              key={index}
              onClick={() => clickHandler(s.name)}
              provider={s.name || ''}
              data-test-id="frontier-sdk-signup-page-oidc-btn"
            ></OIDCButton>
          );
        })}

        {mailotp && <MagicLink />}
      </Flex>
      <div style={{ fontWeight: '400' }}>
        <Text size="small">
          Already have an account?{' '}
          <Link
            href={config.redirectLogin || ''}
            className={styles.redirectLink}
            data-test-id="frontier-sdk-login-btn"
          >
            Login
          </Link>
        </Text>
      </div>
    </Container>
  );
};
