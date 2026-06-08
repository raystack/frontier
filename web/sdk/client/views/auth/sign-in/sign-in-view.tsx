import { Link, Text, Flex } from '@raystack/apsara';
import { ComponentPropsWithRef, ReactNode, useCallback } from 'react';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { useFrontier } from '~/client/contexts/FrontierContext';
import {
  AuthContainer,
  type AuthContainerProps
} from '~/client/components/auth-container';
import { AuthHeader } from '~/client/components/auth-header';
import { AuthOIDCButton } from '~/client/components/auth-oidc-button';
import { MagicLinkView } from '../magic-link/magic-link-view';
import styles from './sign-in-view.module.css';

export type SignInViewProps = ComponentPropsWithRef<'div'> &
  AuthContainerProps & {
    logo?: ReactNode;
    title?: string;
    excludes?: string[];
    footer?: boolean;
  };

export const SignInView = ({
  logo,
  title = 'Login to Raystack',
  excludes = [],
  footer = true,
  ...props
}: SignInViewProps) => {
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
      <Flex direction="column" width="full" gap={5}>
        {filteredOIDC.map((s, index) => {
          return (
            <AuthOIDCButton
              key={index}
              onClick={() => clickHandler(s.name)}
              provider={s.name || ''}
              data-test-id="frontier-sdk-oidc-btn"
            />
          );
        })}

        {mailotp && <MagicLinkView inline />}
      </Flex>
      {footer && (
        <Text size="small" weight="regular">
          Don&apos;t have an account?{' '}
          <Link
            href={config.redirectSignup || ''}
            className={styles.redirectLink}
            data-test-id="frontier-sdk-signup-btn"
          >
            Signup
          </Link>
        </Text>
      )}
    </AuthContainer>
  );
};
