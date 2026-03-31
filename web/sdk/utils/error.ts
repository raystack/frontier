import { Code, ConnectError } from '@connectrpc/connect';

type ErrorHandler = (error: ConnectError) => void;

type ErrorHandlerMap = {
  AlreadyExists?: ErrorHandler;
  InvalidArgument?: ErrorHandler;
  PermissionDenied?: ErrorHandler;
  NotFound?: ErrorHandler;
  Default?: ErrorHandler;
};

const DEFAULT_ERROR_MESSAGE = 'Something went wrong';

export function handleConnectError(
  error: unknown,
  handlers?: ErrorHandlerMap
): void {
  const connectError =
    error instanceof ConnectError
      ? error
      : new ConnectError(
          error instanceof Error ? error.message : DEFAULT_ERROR_MESSAGE
        );

  const defaultHandler =
    handlers?.Default ?? ((err: ConnectError) => console.error(err.message));

  switch (connectError.code) {
    case Code.AlreadyExists:
      handlers?.AlreadyExists
        ? handlers.AlreadyExists(connectError)
        : defaultHandler(connectError);
      break;
    case Code.InvalidArgument:
      handlers?.InvalidArgument
        ? handlers.InvalidArgument(connectError)
        : defaultHandler(connectError);
      break;
    case Code.PermissionDenied:
      handlers?.PermissionDenied
        ? handlers.PermissionDenied(connectError)
        : defaultHandler(connectError);
      break;
    case Code.NotFound:
      handlers?.NotFound
        ? handlers.NotFound(connectError)
        : defaultHandler(connectError);
      break;
    default:
      defaultHandler(connectError);
  }
}
