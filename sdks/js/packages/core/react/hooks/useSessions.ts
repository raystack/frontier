import { useFrontier } from '../contexts/FrontierContext';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';

export interface SessionData {
  id: string;
  browser: string;
  operatingSystem: string;
  ipAddress: string;
  location: string;
  lastActive: string;
  isCurrent: boolean;
}

export const useSessions = () => {
  const { client } = useFrontier();
  const queryClient = useQueryClient();

  const { 
    data: sessionsData, 
    isLoading, 
    error 
  } = useQuery(
    FrontierServiceQueries.listSessions,
    {},
    {
      enabled: !!client,
    }
  );

  const sessions: SessionData[] = (sessionsData?.sessions || []).map((session: any) => ({
    id: session.id || '',
    browser: session.metadata?.browser || 'Unknown',
    operatingSystem: session.metadata?.operatingSystem || 'Unknown',
    ipAddress: session.metadata?.ipAddress || 'Unknown',
    location: session.metadata?.location || 'Unknown',
    lastActive: session.updatedAt?.seconds ? new Date(Number(session.updatedAt.seconds) * 1000).toLocaleString() : 'Unknown',
    isCurrent: session.isCurrent || false,
  }));

  const {
    mutate: revokeSession,
    isPending: isRevokingSession,
  } = useMutation(FrontierServiceQueries.revokeSession, {
    onSuccess: () => {
      // Invalidate and refetch the sessions list
      queryClient.invalidateQueries({
        queryKey: [FrontierServiceQueries.listSessions],
      });
    },
    onError: (error) => {
      console.error('Failed to revoke session:', error);
    },
  });

  const handleRevokeSession = (sessionId: string) => {
    revokeSession({ sessionId });
  };

  return {
    sessions,
    isLoading,
    error: error?.message || null,
    refetch: () => {},
    revokeSession: handleRevokeSession,
    isRevokingSession,
  };
};
