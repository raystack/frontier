import { useState, useEffect } from 'react';
import { useNavigate, useSearch } from '@tanstack/react-router';
import {
  Button,
  toast,
  Text,
  Dialog,
  Flex,
  List,
  Skeleton
} from '@raystack/apsara/v1';
import { useSessions, SessionData } from '../../../hooks/useSessions';
import { RevokeSessionFinalConfirm } from './revoke-session-final-confirm';
import styles from './sessions.module.css';

export const RevokeSessionConfirm = () => {
  const navigate = useNavigate({ from: '/sessions/revoke' });
  const search = useSearch({ from: '/sessions/revoke' }) as { sessionId?: string };
  const { sessions, revokeSession, isRevokingSession } = useSessions();
  const [sessionData, setSessionData] = useState<SessionData | null>(null);
  const [isFinalConfirmOpen, setIsFinalConfirmOpen] = useState(false);

  // Find the session data based on sessionId from URL params
  useEffect(() => {
    if (search.sessionId && sessions.length > 0) {
      const session = sessions.find(s => s.id === search.sessionId);
      if (session) {
        setSessionData(session);
        console.log('Found session for revoking:', session);
      } else {
        console.log('Session not found for ID:', search.sessionId);
      }
    }
  }, [search.sessionId, sessions]);

  const handleRevokeClick = () => {
    setIsFinalConfirmOpen(true);
  };

  const handleFinalConfirm = () => {
    if (!search.sessionId) return;
    
    revokeSession(search.sessionId);
    navigate({ to: '/sessions' });
  };


  return (
    <>
      {sessionData ? (
        <Dialog open={true} onOpenChange={() => navigate({ to: '/sessions' })}>
          <Dialog.Content
            style={{ padding: 0, maxWidth: '400px', width: '100%' }}
          >
            <Dialog.Header className={styles.revokeSessionConfirmHeader}>
              <Flex justify="between" align="center" style={{ width: '100%' }}>
                <Text size="regular">
                  {sessionData.browser} on {sessionData.operatingSystem}
                </Text>
                <Dialog.CloseButton data-test-id="frontier-sdk-close-revoke-session-dialog" />
              </Flex>
            </Dialog.Header>

            <Dialog.Body className={styles.revokeSessionConfirmBody}>
              <List.Root className={styles.listRoot}>
                <List.Item className={styles.listItem}>
                  <List.Label minWidth="120px">Device</List.Label>
                  <List.Value>{sessionData.browser} on {sessionData.operatingSystem}</List.Value>
                </List.Item>
                <List.Item className={styles.listItem}>
                  <List.Label minWidth="120px">IP Address</List.Label>
                  <List.Value>{sessionData.ipAddress}</List.Value>
                </List.Item>
                <List.Item className={styles.listItem}>
                  <List.Label minWidth="120px">Last Location</List.Label>
                  <List.Value>{sessionData.location}</List.Value>
                </List.Item>
                <List.Item className={styles.listItem}>
                  <List.Label minWidth="120px">Last Active</List.Label>
                  <List.Value>{sessionData.lastActive}</List.Value>
                </List.Item>
              </List.Root>
            </Dialog.Body>

            <Dialog.Footer>
              <Flex justify="end" gap={5}>
                <Button
                  variant="outline"
                  color="neutral"
                  onClick={() => navigate({ to: '/sessions' })}
                  data-test-id="frontier-sdk-cancel-revoke-session-dialog"
                >
                  Cancel
                </Button>
                <Button
                  variant="solid"
                  color="danger"
                  onClick={handleRevokeClick}
                  data-test-id="frontier-sdk-confirm-revoke-session-dialog"
                >
                  {sessionData?.isCurrent ? 'Sign out' : 'Revoke'}
                </Button>
              </Flex>
            </Dialog.Footer>
          </Dialog.Content>
        </Dialog>
      ) : (
        <Dialog open={true} onOpenChange={() => navigate({ to: '/sessions' })}>
          <Dialog.Content style={{ padding: 0, maxWidth: '400px', width: '100%' }}>
            <Skeleton 
              height="20px" 
              containerStyle={{ padding: '2rem' }}
              count={4}
            />
          </Dialog.Content>
        </Dialog>
      )}

      <RevokeSessionFinalConfirm
        isOpen={isFinalConfirmOpen}
        onOpenChange={setIsFinalConfirmOpen}
        onConfirm={handleFinalConfirm}
        isLoading={isRevokingSession}
        isCurrentSession={sessionData?.isCurrent}
      />
    </>
  );
};
