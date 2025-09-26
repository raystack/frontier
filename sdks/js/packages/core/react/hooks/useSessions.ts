import { useFrontier } from '../contexts/FrontierContext';
import { useQuery, useMutation } from '@connectrpc/connect-query';
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

  console.log('Using ListSessions API with FrontierServiceQueries...');

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

  console.log('Raw API response:', sessionsData);
  console.log('Loading state:', isLoading);
  console.log('Error state:', error);

  // Map the API response to our SessionData interface
  const sessions: SessionData[] = (sessionsData?.sessions || []).map((session: any) => ({
    id: session.id || '',
    browser: session.metadata?.browser || 'Unknown Browser',
    operatingSystem: session.metadata?.operatingSystem || 'Unknown OS',
    ipAddress: session.metadata?.ipAddress || 'Unknown',
    location: session.metadata?.location || 'Unknown location',
    lastActive: session.updatedAt ? new Date(session.updatedAt.seconds * 1000).toLocaleString() : 'Unknown',
    isCurrent: session.isCurrent || false,
  }));

  console.log('Mapped sessions data:', sessions);

  // RevokeSession mutation
  const {
    mutate: revokeSession,
    isPending: isRevokingSession,
  } = useMutation(FrontierServiceQueries.revokeSession, {
    onSuccess: () => {
      console.log('Session revoked successfully');
      // Refetch sessions after successful revocation
      // Note: useQuery will automatically refetch when the component re-renders
    },
    onError: (error) => {
      console.error('Failed to revoke session:', error);
    },
  });

  const handleRevokeSession = (sessionId: string) => {
    console.log('Revoking session:', sessionId);
    revokeSession({ sessionId });
  };

  return {
    sessions,
    isLoading,
    error: error?.message || null,
    refetch: () => {}, // useQuery handles refetching automatically
    revokeSession: handleRevokeSession,
    isRevokingSession,
  };
};
