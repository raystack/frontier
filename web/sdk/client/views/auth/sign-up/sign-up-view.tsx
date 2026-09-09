import { Field, Flex, Link, Text } from '@raystack/apsara';
import {
  ComponentPropsWithRef,
  ReactNode,
  useCallback,
  useMemo,
  useState
} from 'react';
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
import { AuthConsent, type ConsentLabel } from './auth-consent';
import styles from './sign-up-view.module.css';

export type SignUpViewProps = ComponentPropsWithRef<'div'> &
  AuthContainerProps & {
    logo?: ReactNode;
    title?: string;
    excludes?: string[];
    consentLabel?: ConsentLabel;
    error?: AuthRejection;
  };

export const SignUpView = ({
  logo,
  title = 'Create your account',
  excludes = [],
  consentLabel,
  error,
  ...props
}: SignUpViewProps) => {
  const { config } = useFrontier();
  const [consented, setConsented] = useState(false);
  const [authError, setAuthError] = useState<AuthRejection | null>(() =>
    error?.message ? error : null
  );

  const { data: strategiesData } = useQuery(
    FrontierServiceQueries.listAuthStrategies
  );
  const strategies = (strategiesData?.strategies || [])
    .filter(s => !excludes.includes(s.name))
    .sort(
      (a, b) => Number(a.name === 'mailotp') - Number(b.name === 'mailotp')
    );

  const { data: documents = [], isPending: consentPending } = useQuery(
    FrontierServiceQueries.listConsentDocuments,
    {},
    { select: data => data.documents }
  );

  const acceptedDocumentIds = useMemo(
    () => documents.map(document => document.id),
    [documents]
  );

  const { mutateAsync: authenticate } = useMutation(
    FrontierServiceQueries.authenticate
  );

  const blocked = consentPending || (documents.length > 0 && !consented);

  const clearError = useCallback(() => setAuthError(null), []);

  const clickHandler = useCallback(
    async (name: string) => {
      clearError();
      try {
        const response = await authenticate({
          strategyName: name,
          callbackUrl: config.callbackUrl,
          flowIntent: FlowIntent.SIGNUP,
          acceptedDocumentIds
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
    [authenticate, config, acceptedDocumentIds, clearError]
  );

  const errorFor = (name: string) =>
    authError?.strategy === name ? authError.message : undefined;
  const groupError =
    authError && !strategies.some(s => s.name === authError.strategy)
      ? authError.message
      : undefined;

  return (
    <AuthContainer {...props}>
      <AuthHeader logo={logo} title={title} />
      <Flex
        direction="column"
        gap={5}
        style={{ height: '100%', width: '100%' }}
      >
        {strategies.map(s =>
          s.name === 'mailotp' ? (
            <MagicLinkView
              key={s.name}
              inline
              intent={FlowIntent.SIGNUP}
              acceptedDocumentIds={acceptedDocumentIds}
              disabled={blocked}
              error={errorFor(s.name)}
              onActivate={clearError}
            />
          ) : (
            <Field key={s.name} error={errorFor(s.name)}>
              <AuthOIDCButton
                onClick={() => clickHandler(s.name)}
                provider={s.name}
                disabled={blocked}
                data-test-id="frontier-sdk-signup-page-oidc-btn"
              />
            </Field>
          )
        )}
        {groupError && <Field error={groupError} />}

        {documents.length > 0 && (
          <AuthConsent
            documents={documents}
            checked={consented}
            onCheckedChange={setConsented}
            label={consentLabel}
          />
        )}
      </Flex>
      <Text size="small" weight="regular">
        Already have an account?{' '}
        <Link
          href={config.redirectLogin || ''}
          className={styles.redirectLink}
          data-test-id="frontier-sdk-login-btn"
        >
          Login
        </Link>
      </Text>
    </AuthContainer>
  );
};
