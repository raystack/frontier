import { Link, Text, Flex } from '@raystack/apsara/v1';
import React, { ComponentPropsWithRef, useCallback } from 'react';
import { useFrontier } from '~/react/contexts/FrontierContext';
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
  const { config, client, strategies = [] } = useFrontier();

  const clickHandler = useCallback(
    async (name?: string) => {
      if (!name) return;
      if (!client) return;
      const {
        data: { endpoint = '' }
      } = await client.frontierServiceAuthenticate(name, {
        callback_url: config.callbackUrl
      });
      window.location.href = endpoint;
    },
    [client, config.callbackUrl]
  );

  const mailotp = strategies.find(s => s.name === 'mailotp');
  const filteredOIDC = strategies
    .filter(s => s.name !== 'mailotp')
    .filter(s => !excludes.includes(s.name ?? ''));

  return (
    <Container {...props}>
      <Header logo={logo} title={title} />
      <Flex direction="column" style={{ width: '100%', gap: 'var(--pd-16)' }}>
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
