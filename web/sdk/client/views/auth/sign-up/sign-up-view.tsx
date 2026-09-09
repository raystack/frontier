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
  describeAuthError,
  type AuthRejection
} from '~/client/utils/auth-error';
import { MagicLinkView } from '../magic-link/magic-link-view';
import styles from './sign-up-view.module.css';

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
  const [authError, setAuthError] = useState<AuthRejection | null>(null);
  const callbackError = error?.message ? error : undefined;
  const [showCallbackError, setShowCallbackError] = useState(true);

  const { data: strategiesData } = useQuery(
    FrontierServiceQueries.listAuthStrategies
  );
  const strategies = strategiesData?.strategies || [];

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

  const blocked = consentPending || (documents.length > 0 && !consented);

  const consentContent =
    typeof consentLabel === 'function'
      ? consentLabel(documents)
      : consentLabel ?? defaultConsentLabel(documents);

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
      <Flex
        direction="column"
        gap={5}
        style={{ height: '100%', width: '100%' }}
      >
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
