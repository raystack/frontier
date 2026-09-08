import { Code, ConnectError } from '@connectrpc/connect';

// Both auth RPCs answer a rejection with a connect code and a bare sentinel
// message. FailedPrecondition covers a consent rejection and an account that
// cannot use the requested strategy, so only the message separates those two.
const CONSENT_REQUIRED_MESSAGE =
  'consent required for the configured documents';

export type AuthErrorKind =
  | 'login_user_not_found'
  | 'signup_user_exists'
  | 'consent_required'
  // the request itself was refused: a malformed email, or a bad or expired code
  | 'invalid_request'
  | 'unknown';

export type AuthError = {
  kind: AuthErrorKind;
  message: string;
  // The strategy the rejection came through, read from the response header
  // the callback sets. The flow row knows it when the client does not: an
  // oidc callback carries no strategy name. Absent from older servers.
  strategy?: string;
};

const AUTH_STRATEGY_HEADER = 'frontier-auth-strategy';

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

// A kind that survived a redirect arrives as an unvalidated string, so it is
// checked against the closed set before anything renders it. That is what keeps
// a crafted value out of the page: the copy is this module's, never the URL's.
export const isAuthErrorKind = (value: unknown): value is AuthErrorKind =>
  typeof value === 'string' &&
  AUTH_ERROR_KINDS.includes(value as AuthErrorKind);

export const authErrorMessage = (kind: AuthErrorKind): string =>
  AUTH_ERROR_MESSAGES[kind];

export const describeAuthError = (error: unknown): AuthError => {
  const { code, rawMessage, metadata } = ConnectError.from(error);
  const strategy = metadata.get(AUTH_STRATEGY_HEADER) ?? undefined;

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
