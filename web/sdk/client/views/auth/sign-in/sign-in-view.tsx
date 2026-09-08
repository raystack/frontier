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
  authErrorMessage,
  describeAuthError,
  isAuthErrorKind,
  type AuthErrorKind
} from '~/client/components/auth-error';
import { MagicLinkView } from '../magic-link/magic-link-view';
import styles from './sign-in-view.module.css';

export type SignInViewProps = ComponentPropsWithRef<'div'> &
  AuthContainerProps & {
    logo?: ReactNode;
    title?: string;
    excludes?: string[];
    footer?: boolean;
    // A rejection that arrived at the application's callback page, for oidc and
    // mail link, where the email is unknown before the redirect. How it travels
    // back here is the application's business: this view only renders it.
    error?: AuthErrorKind;
    // The strategy that rejection came through: the callback names it on the
    // error, and the application carries it here with the kind. A rejection
    // with a strategy attaches to that button; without one it is a message for
    // the whole group, since guessing a button would be a lie.
    errorStrategy?: string;
  };

type StrategyRejection = {
  // undefined when the rejection cannot be tied to one strategy
  strategy?: string;
  message: string;
};

export const SignInView = ({
  logo,
  title = 'Login to Raystack',
  excludes = [],
  footer = true,
  error,
  errorStrategy,
  ...props
}: SignInViewProps) => {
  const { config } = useFrontier();
  const [authError, setAuthError] = useState<StrategyRejection | null>(null);
  // The prop crossed a redirect as an unvalidated string, so an unknown value
  // is dropped rather than rendered.
  const callbackError = isAuthErrorKind(error) ? error : undefined;
  // Any click here supersedes what the page arrived with.
  const [showCallbackError, setShowCallbackError] = useState(true);

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
      setAuthError(null);
      setShowCallbackError(false);
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
    [authenticate, config]
  );

  const mailotp = strategies.find(s => s.name === 'mailotp');
  const filteredOIDC = strategies
    .filter(s => s.name !== 'mailotp')
    .filter(s => !excludes.includes(s.name ?? ''));

  const rejection =
    authError ??
    (showCallbackError && callbackError
      ? { strategy: errorStrategy, message: authErrorMessage(callbackError) }
      : undefined);

  // A rejection whose strategy is not one of the buttons rendered here has
  // nowhere to attach, so it reads as a message for the group instead.
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
        {/* One Field per button, so a rejection sits under the strategy that
            caused it. The email form stays outside: wrapping it would nest a
            second Field inside this one. */}
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
