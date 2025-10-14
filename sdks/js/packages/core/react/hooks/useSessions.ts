import { useMemo } from 'react';
import { useQuery, useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { toast } from '@raystack/apsara';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { timestampToDayjs } from '../../utils/timestamp';

dayjs.extend(relativeTime);

// Utility function to format error messages based on status code
const getErrorMessage = (error: unknown): string => {
  if (error instanceof Error && 'status' in error && error.status === 500) {
    return 'Something went wrong';
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'Something went wrong';
};

export const formatDeviceDisplay = (browser?: string, operatingSystem?: string): string => {
  const browserName = browser || "Unknown";
  const osName = operatingSystem || "Unknown";
  return browserName === "Unknown" && osName === "Unknown" ? "Unknown browser and OS" : `${browserName} on ${osName}`;
};

export const useSessions = () => {
  const queryClient = useQueryClient();
  const transport = useTransport();

  const { 
    data: sessionsData, 
    isLoading, 
    error 
  } = useQuery(
    FrontierServiceQueries.listSessions,
    {}
  );

  const formatLastActive = (updatedAt?: any) => {
    const d = timestampToDayjs(updatedAt);
    return d ? d.fromNow() : "Unknown";
  };

  const sessions = useMemo(() => 
    (sessionsData?.sessions || [])
      .map((session: any) => ({
        id: session.id || '',
        browser: session.metadata?.browser || 'Unknown',
        operatingSystem: session.metadata?.operatingSystem || 'Unknown',
        ipAddress: session.metadata?.ipAddress || 'Unknown',
        location: session.metadata?.location || 'Unknown',
        lastActive: formatLastActive(session.updatedAt),
        isCurrent: session.isCurrentSession || false,
      }))
      .sort((a, b) => {
        // Current session first, then by last active (most recent first)
        if (a.isCurrent && !b.isCurrent) return -1;
        if (!a.isCurrent && b.isCurrent) return 1;
        return 0; // Keep original order for non-current sessions
      }), [sessionsData?.sessions]
  );

  const {
    mutate: revokeSession,
    isPending: isRevokingSession,
  } = useMutation(FrontierServiceQueries.revokeSession, {
    onSuccess: () => {
      // Invalidate and refetch the sessions list
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: FrontierServiceQueries.listSessions,
          transport,
          input: {},
          cardinality: "finite",
        }),
      });
      toast.success('Session revoked successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to revoke session', {
        description: getErrorMessage(error)
      });
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
