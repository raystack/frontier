import { Field, Flex, Link, Text } from '@raystack/apsara';
import { ComponentPropsWithRef, ReactNode, useCallback, useState } from 'react';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { FlowIntent, FrontierServiceQueries } from '@raystack/proton/frontier';
import { useFrontier } from '~/client/contexts/FrontierContext';
import {
  AuthContainer,
  type AuthContainerProps
} from '~/client/components/auth-container';
import { AuthHeader } from '~/client/components/auth-header';
import { AuthOIDCButton } from '~/client/components/auth-oidc-button';
import {
  describeAuthError,
  type AuthRejection
} from '~/client/utils/auth-error';
import { MagicLinkView } from '../magic-link/magic-link-view';
import styles from './sign-in-view.module.css';

export type SignInViewProps = ComponentPropsWithRef<'div'> &
  AuthContainerProps & {
    logo?: ReactNode;
    title?: string;
    excludes?: string[];
    footer?: boolean;
    error?: AuthRejection;
  };

export const SignInView = ({
  logo,
  title = 'Login to Raystack',
  excludes = [],
  footer = true,
  error,
  ...props
}: SignInViewProps) => {
  const { config } = useFrontier();
  const [authError, setAuthError] = useState<AuthRejection | null>(null);
  const callbackError = error?.message ? error : undefined;
  const [showCallbackError, setShowCallbackError] = useState(true);

  const { data: strategiesData } = useQuery(
    FrontierServiceQueries.listAuthStrategies
  );
  const strategies = strategiesData?.strategies || [];

  const { mutateAsync: authenticate } = useMutation(
    FrontierServiceQueries.authenticate
  );

  const dismissRejection = useCallback(() => {
    setAuthError(null);
    setShowCallbackError(false);
  }, []);

  const clickHandler = useCallback(
    async (name?: string) => {
      if (!name) return;
      dismissRejection();
      try {
        const response = await authenticate({
          strategyName: name,
          callbackUrl: config.callbackUrl,
          flowIntent: FlowIntent.LOGIN
        });
        if (response.endpoint) {
          window.location.href = response.endpoint;
        }
      } catch (err) {
        setAuthError({
          strategy: name,
          message: describeAuthError(err).message
        });
      }
    },
    [authenticate, config, dismissRejection]
  );

  const mailotp = strategies.find(s => s.name === 'mailotp');
  const filteredOIDC = strategies
    .filter(s => s.name !== 'mailotp')
    .filter(s => !excludes.includes(s.name ?? ''));

  const rejection =
    authError ?? (showCallbackError ? callbackError : undefined);

  const mailRejection =
    mailotp && rejection?.strategy === 'mailotp'
      ? rejection.message
      : undefined;
  const attached =
    !!mailRejection ||
    filteredOIDC.some(s => s.name && s.name === rejection?.strategy);
  const groupMessage = rejection && !attached ? rejection.message : undefined;

  return (
    <AuthContainer {...props}>
      <AuthHeader logo={logo} title={title} />
      <Flex direction="column" gap={5} style={{ width: '100%' }}>
        <Flex direction="column" gap={5}>
          {filteredOIDC.map((s, index) => {
            return (
              <Field
                key={index}
                error={
                  s.name && s.name === rejection?.strategy
                    ? rejection.message
                    : undefined
                }
              >
                <AuthOIDCButton
                  onClick={() => clickHandler(s.name)}
                  provider={s.name || ''}
                  data-test-id="frontier-sdk-oidc-btn"
                />
              </Field>
            );
          })}
        </Flex>

        {mailotp && (
          <MagicLinkView
            inline
            intent={FlowIntent.LOGIN}
            error={mailRejection}
            onActivate={dismissRejection}
          />
        )}
        {groupMessage && <Field error={groupMessage} />}
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
