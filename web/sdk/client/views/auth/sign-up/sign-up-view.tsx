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
import { visibleAuthStrategies } from '~/client/utils/auth-strategies';
import { MAIL_OTP_STRATEGY } from '~/client/utils/constants';
import { MagicLinkView } from '../magic-link/magic-link-view';
import { AuthConsent, type ConsentLabel } from './auth-consent';
import styles from './sign-up-view.module.css';

const CONSENT_UNAVAILABLE_MESSAGE =
  'The documents required to sign up could not be loaded. Please refresh the page.';
const CONSENT_REQUIRED_TOOLTIP = 'You must agree to continue';

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
  const [agreedDocumentIds, setAgreedDocumentIds] = useState<string[]>([]);
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

  const {
    data: documents = [],
    isPending: consentPending,
    isError: consentFailed
  } = useQuery(
    FrontierServiceQueries.listConsentDocuments,
    {},
    { select: data => data.documents }
  );

  const consented =
    documents.length === agreedDocumentIds.length &&
    documents.every(document => agreedDocumentIds.includes(document.id));
  const acceptedDocumentIds = useMemo(
    () => (consented ? agreedDocumentIds : []),
    [consented, agreedDocumentIds]
  );

  const { mutateAsync: authenticate } = useMutation(
    FrontierServiceQueries.authenticate
  );

  const blocked =
    consentPending || consentFailed || (documents.length > 0 && !consented);

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

  const isUnknownError =
    authError !== null &&
    !strategies.some(
      s => s.name !== MAIL_OTP_STRATEGY && s.name === authError.strategy
    );

  return (
    <AuthContainer {...props}>
      <AuthHeader logo={logo} title={title} />
      <Flex
        direction="column"
        gap={5}
        style={{ height: '100%', width: '100%' }}
      >
        {strategies.map(s =>
          s.name === MAIL_OTP_STRATEGY ? (
            <MagicLinkView
              key={s.name}
              inline
              intent={FlowIntent.SIGNUP}
              acceptedDocumentIds={acceptedDocumentIds}
              disabled={blocked}
              disabledMessage={CONSENT_REQUIRED_TOOLTIP}
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
                disabled={blocked}
                disabledMessage={CONSENT_REQUIRED_TOOLTIP}
                data-test-id="frontier-sdk-signup-page-oidc-btn"
              />
            </Field>
          )
        )}
        {(consentFailed || isUnknownError) && (
          <Field
            error={
              consentFailed ? CONSENT_UNAVAILABLE_MESSAGE : authError?.message
            }
          />
        )}
        {documents.length > 0 && (
          <AuthConsent
            documents={documents}
            checked={consented}
            onCheckedChange={checked =>
              setAgreedDocumentIds(
                checked ? documents.map(document => document.id) : []
              )
            }
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
