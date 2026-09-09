import { Code, ConnectError } from '@connectrpc/connect';
import { AuthStrategySchema } from '@raystack/proton/frontier';

const SERVER_CONSENT_REQUIRED = 'consent required for the configured documents';

export type AuthRejection = {
  message: string;
  strategy?: string;
};

const LOGIN_USER_NOT_FOUND_MESSAGE =
  'No account found for this email. Please sign up to create an account.';
const SIGNUP_USER_EXISTS_MESSAGE =
  'An account already exists for this email. Please log in to continue.';
const CONSENT_REQUIRED_MESSAGE =
  'Please accept the required documents to continue.';
const UNKNOWN_MESSAGE = 'Something went wrong. Please try again.';

const readStrategy = (error: ConnectError): string | undefined => {
  const [strategy] = error.findDetails(AuthStrategySchema);
  return strategy?.name || undefined;
};

const messageFor = (code: Code, rawMessage: string): string => {
  switch (code) {
    case Code.NotFound:
      return LOGIN_USER_NOT_FOUND_MESSAGE;
    case Code.AlreadyExists:
      return SIGNUP_USER_EXISTS_MESSAGE;
    case Code.FailedPrecondition:
      return rawMessage.includes(SERVER_CONSENT_REQUIRED)
        ? CONSENT_REQUIRED_MESSAGE
        : rawMessage;
    case Code.InvalidArgument:
      return rawMessage;
    default:
      return UNKNOWN_MESSAGE;
  }
};

export const describeAuthError = (error: unknown): AuthRejection => {
  const connectError = ConnectError.from(error);
  return {
    message: messageFor(connectError.code, connectError.rawMessage),
    strategy: readStrategy(connectError)
  };
};
