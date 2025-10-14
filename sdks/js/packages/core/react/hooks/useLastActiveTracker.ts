import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';

export const useLastActiveTracker = () => {
  useQuery(FrontierServiceQueries.pingUserSession, {}, {
    // Ping immediately and then every 10 minutes
    refetchInterval: 10 * 60 * 1000,
    refetchIntervalInBackground: true,
    staleTime: Infinity,
    gcTime: Infinity,
    retry: false,
  });
};