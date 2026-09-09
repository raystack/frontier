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
import { visibleAuthStrategies } from '~/client/utils/auth-strategies';
import { MAIL_OTP_STRATEGY } from '~/client/utils/constants';
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
  const [authError, setAuthError] = useState<AuthRejection | null>(() =>
    error?.message ? error : null
  );

  const { data: strategiesData } = useQuery(
    FrontierServiceQueries.listAuthStrategies
  );
  const strategies = visibleAuthStrategies(
    strategiesData?.strategies || [],
    excludes
  );

  const { mutateAsync: authenticate } = useMutation(
    FrontierServiceQueries.authenticate
  );

  const clearError = useCallback(() => setAuthError(null), []);

  const clickHandler = useCallback(
    async (name: string) => {
      clearError();
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
    [authenticate, config, clearError]
  );

  const isUnknownError =
    authError !== null &&
    !strategies.some(
      s => s.name !== MAIL_OTP_STRATEGY && s.name === authError.strategy
    );

  return (
    <AuthContainer {...props}>
      <AuthHeader logo={logo} title={title} />
      <Flex direction="column" gap={5} style={{ width: '100%' }}>
        {strategies.map(s =>
          s.name === MAIL_OTP_STRATEGY ? (
            <MagicLinkView
              key={s.name}
              inline
              intent={FlowIntent.LOGIN}
              onActivate={clearError}
            />
          ) : (
            <Field
              key={s.name}
              error={
                authError?.strategy === s.name ? authError.message : undefined
              }
            >
              <AuthOIDCButton
                onClick={() => clickHandler(s.name)}
                provider={s.name}
                data-test-id="frontier-sdk-oidc-btn"
              />
            </Field>
          )
        )}
        {isUnknownError && <Field error={authError.message} />}
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
