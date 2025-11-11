import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, PingUserSessionResponse } from '@raystack/proton/frontier';

export const useLastActiveTracker = ({ enabled = false })
: { data: PingUserSessionResponse | undefined, isLoading: boolean, error: Error | undefined } => {
  const { data, isLoading, error } = useQuery(
    FrontierServiceQueries.pingUserSession,
    {},
    {
      enabled: enabled,
      // Ping immediately and then every 10 minutes
      refetchInterval: 10 * 60 * 1000,
      refetchIntervalInBackground: true,
      staleTime: Infinity,
      gcTime: Infinity,
      retry: false
    }
  );
  return { data: data, isLoading, error: error ? new Error(error.message) : undefined };
};
