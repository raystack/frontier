import { Checkbox, Field, Flex, Link, Text } from '@raystack/apsara';
import {
  ComponentPropsWithRef,
  Fragment,
  ReactNode,
  useCallback,
  useMemo,
  useState
} from 'react';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import {
  FlowIntent,
  FrontierServiceQueries,
  type ConsentDocument
} from '@raystack/proton/frontier';
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
import styles from './sign-up-view.module.css';

// The consent label: static copy, or a function of the documents for copy that
// links them. The default names every document with a link to its url.
export type ConsentLabel =
  | ReactNode
  | ((documents: ConsentDocument[]) => ReactNode);

const documentLinks = (documents: ConsentDocument[]) =>
  documents.map((document, index) => (
    <Fragment key={document.id}>
      {index > 0 && (index === documents.length - 1 ? ' and ' : ', ')}
      <Link
        href={document.url}
        external
        variant="primary"
        size="micro"
        data-test-id={`frontier-sdk-consent-document-${document.id}`}
      >
        {document.title}
      </Link>
    </Fragment>
  ));

const defaultConsentLabel = (documents: ConsentDocument[]) => (
  <>I agree to the {documentLinks(documents)}</>
);

export type SignUpViewProps = ComponentPropsWithRef<'div'> &
  AuthContainerProps & {
    logo?: ReactNode;
    title?: string;
    excludes?: string[];
    consentLabel?: ConsentLabel;
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

export const SignUpView = ({
  logo,
  title = 'Create your account',
  excludes = [],
  consentLabel,
  error,
  errorStrategy,
  ...props
}: SignUpViewProps) => {
  const { config } = useFrontier();
  const [consented, setConsented] = useState(false);
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

  // An empty list means the deployment asks for no consent, so there is no
  // checkbox and this is the view as it was before.
  const { data: consentData, isPending: consentPending } = useQuery(
    FrontierServiceQueries.listConsentDocuments
  );
  const documents = useMemo(() => consentData?.documents ?? [], [consentData]);

  const acceptedDocumentIds = useMemo(
    () => documents.map(document => document.id),
    [documents]
  );

  const { mutateAsync: authenticate } = useMutation(
    FrontierServiceQueries.authenticate
  );

  // Every sign-up control stays disabled until the documents are accepted, and
  // while the list is in flight: until it resolves there is no way to know
  // whether this deployment asks for consent at all, and a click before it
  // lands would send an empty id set for a deployment that does.
  const blocked = consentPending || (documents.length > 0 && !consented);

  const consentContent =
    typeof consentLabel === 'function'
      ? consentLabel(documents)
      : consentLabel ?? defaultConsentLabel(documents);

  // Starting any strategy supersedes whatever rejection is on screen.
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
    [authenticate, config, acceptedDocumentIds, dismissRejection]
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
      <Flex
        direction="column"
        gap={5}
        style={{ height: '100%', width: '100%' }}
      >
        {/* One Field per button, so a rejection sits under the strategy that
            caused it. The email form and the consent checkbox stay outside:
            the first would nest a second Field, the second would be marked
            invalid for a rejection that is not about it. */}
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
                  disabled={blocked}
                  data-test-id="frontier-sdk-signup-page-oidc-btn"
                />
              </Field>
            );
          })}
        </Flex>

        {mailotp && (
          <MagicLinkView
            inline
            intent={FlowIntent.SIGNUP}
            acceptedDocumentIds={acceptedDocumentIds}
            disabled={blocked}
            error={mailRejection}
            onActivate={dismissRejection}
          />
        )}
        {groupMessage && <Field error={groupMessage} />}

        {documents.length > 0 && (
          <Flex
            gap={4}
            align="start"
            justify="center"
            className={styles.consent}
            render={<label />}
          >
            <Checkbox
              size="small"
              checked={consented}
              onCheckedChange={setConsented}
              data-test-id="frontier-sdk-consent-checkbox"
            />
            <Text size="micro" variant="secondary">
              {consentContent}
            </Text>
          </Flex>
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
