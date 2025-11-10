import { useRef, useState } from 'react';
import { useQuery } from '@connectrpc/connect-query';
import type { ConnectError } from '@connectrpc/connect';
import type { UseQueryResult } from '@tanstack/react-query';
import type {
  DescMessage,
  DescMethodUnary,
  MessageInitShape,
  MessageShape
} from '@bufbuild/protobuf';

type UseConnectQueryPollingResult<TData> = UseQueryResult<
  TData,
  ConnectError
> & {
  isPollingFailed: boolean;
};

export function useConnectQueryPolling<
  I extends DescMessage,
  O extends DescMessage,
  SelectOutData = MessageShape<O>
>(params: {
  method: DescMethodUnary<I, O>;
  input: MessageInitShape<I> | undefined;
  enabled: boolean;
  select: (data: MessageShape<O>) => SelectOutData;
  isComplete: (data: SelectOutData) => boolean;
  maxPoll?: number;
  pollingInterval?: number;
}): UseConnectQueryPollingResult<SelectOutData> {
  const {
    method,
    input,
    enabled,
    select,
    isComplete,
    maxPoll = 40,
    pollingInterval = 3000
  } = params;
  const [isPollingFailed, setIsPollingFailed] = useState(false);
  const pollCountRef = useRef(0);

  const result = useQuery<I, O, SelectOutData>(method, input, {
    enabled,
    select,
    retry: false,
    refetchInterval: query => {
      if (query.state.error) {
        setIsPollingFailed(true);
        return false;
      }

      if (pollCountRef.current >= maxPoll) {
        setIsPollingFailed(true);
        return false;
      }

      pollCountRef.current++;

      if (query.state.data !== undefined) {
        const transformed = select(query.state.data);
        if (transformed !== undefined && isComplete(transformed)) {
          return false;
        }
      }

      return pollingInterval;
    }
  });

  return {
    ...result,
    isPollingFailed
  } as UseConnectQueryPollingResult<SelectOutData>;
}
