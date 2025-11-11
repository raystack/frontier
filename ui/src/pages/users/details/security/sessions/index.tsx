import { useState } from "react";
import { Flex, Text, Button, toast } from "@raystack/apsara";
import { useUser } from "../../user-context";
import { RevokeSessionConfirm } from "./revoke-session-confirm";
import { SessionSkeleton } from "./session-skeleton";
import { useQuery, useMutation, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { AdminServiceQueries, Session } from "@raystack/proton/frontier";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import { timestampToDate } from "~/utils/connect-timestamp";
import styles from "./sessions.module.css";

/**
 * Formats location from structured object to display string
 * @param location - Location object with city, country, latitude, longitude, or legacy string format
 * @returns Formatted string like "City, Country" or "Unknown" if empty
 * Note: This function also exists in the SDK utils/index.ts file. 
 * If you make any changes here, please update the SDK utils/index.ts file as well.
 */
const formatLocation = (location?: string | {
  city?: string;
  country?: string;
  latitude?: string;
  longitude?: string;
}): string => {
  if (!location) return 'Unknown location';
  
  const city = location.city?.trim() || '';
  const country = location.country?.trim() || '';
  
  if (city && country) {
    return `${city}, ${country}`;
  }
  if (city) {
    return city;
  }
  if (country) {
    return country;
  }
  
  return 'Unknown location';
};

dayjs.extend(relativeTime);

const getErrorMessage = (error: any): string => {
  if (error?.status === 500) {
    return 'Something went wrong';
  }
  return error?.message || 'Something went wrong';
};

export const formatDeviceDisplay = (browser?: string, operatingSystem?: string): string => {
  const browserName = browser || "Unknown";
  const osName = operatingSystem || "Unknown";
  return browserName === "Unknown" && osName === "Unknown" ? "Unknown browser and OS" : `${browserName} on ${osName}`;
};

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
        description: getErrorMessage(error)
      });
    },
  });

  const handleRevoke = (session: Session) => {
    setSelectedSession({
      browser: session.metadata?.browser || "Unknown",
      operatingSystem: session.metadata?.operatingSystem || "Unknown",
      ipAddress: session.metadata?.ipAddress || "Unknown",
      location: formatLocation(session.metadata?.location),
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
    
    const date = timestampToDate(updatedAt);
    if (!date) return "Unknown";
    
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
          <SessionSkeleton count={3} />
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
                    {formatLocation(session.metadata?.location)}
                  </Text>
                  <Text variant="tertiary" size="small">•</Text>
                  {session.isCurrentSession ? (
                    <Text variant="success" size="small">Current session</Text>
                  ) : (
                    <Text variant="tertiary" size="small">
                      Last active {formatLastActive(session.updatedAt)}
                    </Text>
                  )}
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