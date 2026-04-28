import { Link, Text, Flex } from '@raystack/apsara-v1';
import { ComponentPropsWithRef, ReactNode, useCallback } from 'react';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { useFrontier } from '~/react/contexts/FrontierContext';
import {
  AuthContainer,
  type AuthContainerProps
} from '../components/auth-container';
import { AuthHeader } from '../components/auth-header';
import { MagicLinkForm } from '../components/magic-link-form';
import { OIDCButton } from '../components/oidc-button';
import styles from './sign-up-view.module.css';

export type SignUpViewProps = ComponentPropsWithRef<'div'> &
  AuthContainerProps & {
    logo?: ReactNode;
    title?: string;
    excludes?: string[];
  };

export const SignUpView = ({
  logo,
  title = 'Create your account',
  excludes = [],
  ...props
}: SignUpViewProps) => {
  const { config } = useFrontier();

  const { data: strategiesData } = useQuery(
    FrontierServiceQueries.listAuthStrategies
  );
  const strategies = strategiesData?.strategies || [];

  const { mutateAsync: authenticate } = useMutation(
    FrontierServiceQueries.authenticate
  );

  const clickHandler = useCallback(
    async (name?: string) => {
      if (!name) return;
      try {
        const response = await authenticate({
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
    [authenticate, config.callbackUrl]
  );

  const mailotp = strategies.find(s => s.name === 'mailotp');
  const filteredOIDC = strategies
    .filter(s => s.name !== 'mailotp')
    .filter(s => !excludes.includes(s.name ?? ''));

  return (
    <AuthContainer {...props}>
      <AuthHeader logo={logo} title={title} />
      <Flex direction="column" gap={3} width="full">
        {filteredOIDC.map((s, index) => {
          return (
            <OIDCButton
              key={index}
              onClick={() => clickHandler(s.name)}
              provider={s.name || ''}
              data-test-id="frontier-sdk-signup-page-oidc-btn"
            />
          );
        })}

        {mailotp && <MagicLinkForm />}
      </Flex>
      <div className={styles.footer}>
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
    </AuthContainer>
  );
};
