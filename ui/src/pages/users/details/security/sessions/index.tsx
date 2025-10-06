import { useState } from "react";
import { Flex, Text, Button, Skeleton, toast } from "@raystack/apsara/v1";
import { useUser } from "../../user-context";
import { RevokeSessionConfirm } from "./revoke-session-confirm";
import { useQuery, useMutation, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import styles from "./sessions.module.css";

dayjs.extend(relativeTime);

export const formatDeviceDisplay = (browser?: string, operatingSystem?: string): string => {
  const browserName = browser || "Unknown";
  const osName = operatingSystem || "Unknown";
  return browserName === "Unknown" && osName === "Unknown" ? "Unknown browser and OS" : `${browserName} on ${osName}`;
};

interface SessionData {
  id?: string;
  metadata?: {
    operatingSystem?: string;
    browser?: string;
    ipAddress?: string;
    location?: string;
  };
  updatedAt?: {
    seconds: bigint;
  };
}

export const UserSessions = () => {
  const { user } = useUser();
  const queryClient = useQueryClient();
  const transport = useTransport();
  const [isRevokeDialogOpen, setIsRevokeDialogOpen] = useState(false);
  const [selectedSession, setSelectedSession] = useState<{
    browser: string;
    operatingSystem: string;
    ipAddress: string;
    location: string;
    lastActive: string;
    sessionId: string;
  } | null>(null);

  const { 
    data: sessionsData, 
    isLoading, 
    error 
  } = useQuery(
    AdminServiceQueries.listUserSessions,
    { userId: user?.id || "" },
    {
      enabled: !!user?.id,
    }
  );

  const {
    mutate: revokeUserSession,
    isPending: isRevokingSession,
  } = useMutation(AdminServiceQueries.revokeUserSession, {
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: AdminServiceQueries.listUserSessions,
          transport,
          input: { userId: user?.id || "" },
          cardinality: "finite",
        }),
      });
      toast.success('Session revoked successfully');
    },
    onError: (error: any) => {
      toast.error('Failed to revoke session', {
        description: error.message || 'Something went wrong'
      });
    },
  });

  const handleRevoke = (session: SessionData) => {
    setSelectedSession({
      browser: session.metadata?.browser || "Unknown",
      operatingSystem: session.metadata?.operatingSystem || "Unknown",
      ipAddress: session.metadata?.ipAddress || "Unknown",
      location: session.metadata?.location || "Unknown",
      lastActive: formatLastActive(session.updatedAt),
      sessionId: session.id || ""
    });
    setIsRevokeDialogOpen(true);
  };

  const handleRevokeConfirm = () => {
    if (selectedSession?.sessionId) {
      revokeUserSession({ sessionId: selectedSession.sessionId });
    }
  };


  const formatLastActive = (updatedAt?: any) => {
    if (!updatedAt) return "Unknown";
    
    const seconds = typeof updatedAt.seconds === 'bigint' ? Number(updatedAt.seconds) : updatedAt.seconds;
    const date = new Date(seconds * 1000);
    return dayjs(date).fromNow();
  };

  const renderSessionsHeader = () => (
    <Flex direction="column" gap={3}>
      <Text size='large' weight='medium'>Sessions</Text>
      <Text size='regular' variant="secondary">
        Devices logged into this account.
      </Text>
    </Flex>
  );

  if (isLoading) {
    return (
      <Flex direction="column" gap={9}>
        {renderSessionsHeader()}
        <Flex direction="column" className={styles.sessionsContainer}>
          <Skeleton 
            height="32px"
            containerStyle={{ padding: '1rem 0' }}
            count={3}
          />
        </Flex>
      </Flex>
    );
  }

  if (error) {
    return (
      <Flex direction="column" gap={9}>
        {renderSessionsHeader()}
        <Flex justify="center" align="center" style={{ padding: "2rem" }}>
          <Text color="danger">Failed to load sessions</Text>
        </Flex>
      </Flex>
    );
  }

  const sessions = sessionsData?.sessions || [];

  return (
    <Flex direction="column" gap={9}>
      {renderSessionsHeader()}

      <Flex direction="column" className={styles.sessionsContainer}>
        {sessions.length === 0 ? (
          <Flex justify="center" align="center" style={{ padding: "2rem" }}>
            <Text variant="secondary">No active sessions found</Text>
          </Flex>
        ) : (
          sessions.map((session, index) => (
            <Flex key={session.id} justify="between" align="center" className={styles.sessionItem}>
              <Flex direction="column" gap={3}>
                <Text size="regular">
                  {formatDeviceDisplay(session.metadata?.browser, session.metadata?.operatingSystem)}
                </Text>
                <Flex gap={2} align="center">
                  <Text variant="tertiary" size="small">
                    {session.metadata?.location || "Unknown location"}
                  </Text>
                  <Text variant="tertiary" size="small">•</Text>
                  <Text variant="tertiary" size="small">
                    Last active {formatLastActive(session.updatedAt)}
                  </Text>
                </Flex>
              </Flex>
              <Button 
                variant="text" 
                color="neutral" 
                data-test-id={`frontier-ui-revoke-session-${index + 1}`}
                onClick={() => handleRevoke(session)}
              >
                Revoke
              </Button>
            </Flex>
          ))
        )}
      </Flex>

      <RevokeSessionConfirm
        isOpen={isRevokeDialogOpen}
        onOpenChange={setIsRevokeDialogOpen}
        sessionInfo={selectedSession || undefined}
        onRevokeConfirm={handleRevokeConfirm}
        isLoading={isRevokingSession}
      />
    </Flex>
  );
}; 