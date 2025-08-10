import { useState } from "react";
import { Flex, Text, Button } from "@raystack/apsara/v1";
import { useUser } from "../../user-context";
import { RevokeSessionConfirm } from "./revoke-session-confirm";
import styles from "./sessions.module.css";

export const UserSessions = () => {
  const { user } = useUser();
  const [isRevokeDialogOpen, setIsRevokeDialogOpen] = useState(false);
  const [selectedSession, setSelectedSession] = useState<{
    device: string;
    ipAddress: string;
    location: string;
    lastActive: string;
  } | null>(null);

  const handleRevoke = (sessionId: string, sessionInfo: any) => {
    setSelectedSession(sessionInfo);
    setIsRevokeDialogOpen(true);
  };

  return (
    <Flex direction="column" gap={9}>
      <Flex direction="column" gap={3}>
        <Text size='large' weight='medium'>Sessions</Text>
        <Text size='regular' variant="secondary">
            Devices logged into this account.
        </Text>
      </Flex>

      <Flex direction="column" className={styles.sessionsContainer}>
        <Flex justify="between" align="center" className={styles.sessionItem}>
          <Flex direction="column" gap={3}>
            <Text size="regular">Chrome on Mac OS x</Text>
            <Flex gap={2} align="center">
              <Text variant="tertiary" size="small">Bangalore</Text>
              <Text variant="tertiary" size="small">•</Text>
              <Text variant="tertiary" size="small">Last active 10 minutes ago</Text>
            </Flex>
          </Flex>
          <Button variant="text" color="neutral" data-test-id="frontier-ui-revoke-session-1" onClick={() => handleRevoke("session1", {
            device: "Chrome on Mac OS x",
            ipAddress: "203.0.113.25",
            location: "Bangalore, India",
            lastActive: "10 minutes ago"
          })}>
            Revoke
          </Button>
        </Flex>

        <Flex justify="between" align="center" className={styles.sessionItem}>
          <Flex direction="column" gap={3}>
            <Text size="regular">Safari on iPhone</Text>
            <Flex gap={2} align="center">
              <Text variant="tertiary" size="small">Mumbai</Text>
              <Text variant="tertiary" size="small">•</Text>
              <Text variant="tertiary" size="small">Last active 2 hours ago</Text>
            </Flex>
          </Flex>
          <Button variant="text" color="neutral" data-test-id="frontier-ui-revoke-session-2" onClick={() => handleRevoke("session2", {
            device: "Safari on iPhone",
            ipAddress: "198.51.100.42",
            location: "Mumbai, India",
            lastActive: "2 hours ago"
          })}>
            Revoke
          </Button>
        </Flex>
      </Flex>

      <RevokeSessionConfirm
        isOpen={isRevokeDialogOpen}
        onOpenChange={setIsRevokeDialogOpen}
        sessionInfo={selectedSession || undefined}
      />
    </Flex>
  );
}; 