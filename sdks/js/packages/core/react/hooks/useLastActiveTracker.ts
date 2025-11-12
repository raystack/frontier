import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, PingUserSessionResponse } from '@raystack/proton/frontier';

type UseLastActiveTrackerParams = {
  enabled?: boolean;
};

type UseLastActiveTrackerReturn = {
  data: PingUserSessionResponse | undefined;
  isLoading: boolean;
  error: Error | null;
};

export const useLastActiveTracker = ({ enabled = false }: UseLastActiveTrackerParams = {}): UseLastActiveTrackerReturn => {
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
  return { data, isLoading, error };
};
