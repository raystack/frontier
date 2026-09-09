import { Code, ConnectError } from '@connectrpc/connect';
import { AuthStrategySchema } from '@raystack/proton/frontier';

const CONSENT_REQUIRED_MESSAGE =
  'consent required for the configured documents';

export type AuthErrorKind =
  | 'login_user_not_found'
  | 'signup_user_exists'
  | 'consent_required'
  | 'invalid_request'
  | 'unknown';

export type AuthError = {
  kind: AuthErrorKind;
  message: string;
  strategy?: string;
};

const readStrategy = (error: ConnectError): string | undefined => {
  const [strategy] = error.findDetails(AuthStrategySchema);
  return strategy?.name || undefined;
};

const AUTH_ERROR_MESSAGES: Record<AuthErrorKind, string> = {
  login_user_not_found:
    'No account found for this email. Please sign up to create an account.',
  signup_user_exists:
    'An account already exists for this email. Please log in to continue.',
  consent_required: 'Please accept the required documents to continue.',
  invalid_request: 'That request could not be completed. Please try again.',
  unknown: 'Something went wrong. Please try again.'
};

const AUTH_ERROR_KINDS = Object.keys(AUTH_ERROR_MESSAGES) as AuthErrorKind[];

export const isAuthErrorKind = (value: unknown): value is AuthErrorKind =>
  typeof value === 'string' &&
  AUTH_ERROR_KINDS.includes(value as AuthErrorKind);

export const authErrorMessage = (kind: AuthErrorKind): string =>
  AUTH_ERROR_MESSAGES[kind];

export const describeAuthError = (error: unknown): AuthError => {
  const connectError = ConnectError.from(error);
  const { code, rawMessage } = connectError;
  const strategy = readStrategy(connectError);

  switch (code) {
    case Code.NotFound:
      return {
        kind: 'login_user_not_found',
        message: AUTH_ERROR_MESSAGES.login_user_not_found,
        strategy
      };
    case Code.AlreadyExists:
      return {
        kind: 'signup_user_exists',
        message: AUTH_ERROR_MESSAGES.signup_user_exists,
        strategy
      };
    case Code.FailedPrecondition:
      return rawMessage.includes(CONSENT_REQUIRED_MESSAGE)
        ? {
          kind: 'consent_required',
          message: AUTH_ERROR_MESSAGES.consent_required,
          strategy
        }
        : { kind: 'unknown', message: rawMessage, strategy };
    case Code.InvalidArgument:
      return { kind: 'invalid_request', message: rawMessage, strategy };
    default:
      return {
        kind: 'unknown',
        message: AUTH_ERROR_MESSAGES.unknown,
        strategy
      };
  }
};
