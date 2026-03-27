'use client';

import { useState } from 'react';
import { DotFilledIcon } from '@radix-ui/react-icons';
import { Button, Flex, Skeleton, Text } from '@raystack/apsara-v1';
import { useSessions } from '~/react/hooks/useSessions';
import { ViewContainer } from '~/react/components/view-container';
import { ViewHeader } from '~/react/components/view-header';
import { RevokeSessionDialog } from './components/revoke-session-dialog';
import styles from './sessions-view.module.css';

export interface SessionsViewProps {
  onLogout?: () => void;
  onRevoke?: () => void;
}

export function SessionsView({ onLogout, onRevoke }: SessionsViewProps = {}) {
  const { sessions, isLoading, error } = useSessions();

  const [revokeState, setRevokeState] = useState({
    open: false,
    sessionId: ''
  });

  const handleRevokeOpenChange = (value: boolean) => {
    if (!value) {
      setRevokeState({ open: false, sessionId: '' });
    } else {
      setRevokeState(prev => ({ ...prev, open: value }));
    }
  };

  const handleRevoke = (sessionId: string) => {
    setRevokeState({ open: true, sessionId });
  };

  return (
    <ViewContainer>
      <ViewHeader
        title="Sessions"
        description="Devices logged into this account."
      />

      {isLoading ? (
        <Flex direction="column">
          {Array.from({ length: 3 }).map((_, i) => (
            <Flex key={i} justify="between" align="center" className={styles.sessionRow}>
              <Flex direction="column" gap={3}>
                <Skeleton height="24px" width="200px" />
                <Skeleton height="16px" width="140px" />
              </Flex>
              <Skeleton height="32px" width="56px" />
            </Flex>
          ))}
        </Flex>
      ) : error ? (
        <Flex justify="center" align="center" className={styles.stateContainer}>
          <Text size="regular" variant="danger">{error}</Text>
        </Flex>
      ) : sessions.length === 0 ? (
        <Flex justify="center" align="center" className={styles.stateContainer}>
          <Text size="regular" variant="secondary">No active sessions found.</Text>
        </Flex>
      ) : (
        <Flex direction="column">
          {sessions.map(session => (
            <Flex key={session.id} justify="between" align="center" className={styles.sessionRow}>
              <Flex direction="column" gap={3}>
                <Text size="large" weight="medium">
                  {session.browser} on {session.operatingSystem}
                </Text>
                <Flex gap={2} align="center">
                  <Text size="small" variant="tertiary">{session.location}</Text>
                  <DotFilledIcon className={styles.dot} />
                  {session.isCurrent ? (
                    <Text size="small" weight="medium" variant="success">Current session</Text>
                  ) : (
                    <Text size="small" variant="tertiary">Last active {session.lastActive}</Text>
                  )}
                </Flex>
              </Flex>
              <Button
                variant="text"
                color="neutral"
                onClick={() => handleRevoke(session.id)}
                data-test-id="frontier-sdk-revoke-session-button"
              >
                {session.isCurrent ? 'Log out' : 'Revoke'}
              </Button>
            </Flex>
          ))}
        </Flex>
      )}

      <RevokeSessionDialog
        open={revokeState.open}
        onOpenChange={handleRevokeOpenChange}
        sessionId={revokeState.sessionId}
        onLogout={onLogout}
        onRevoke={onRevoke}
      />
    </ViewContainer>
  );
}
