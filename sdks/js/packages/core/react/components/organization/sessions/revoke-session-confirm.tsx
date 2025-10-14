import { useState, useMemo } from 'react';
import { useNavigate, useSearch, useRouteContext } from '@tanstack/react-router';
import {
  Button,
  Text,
  Dialog,
  Flex,
  List,
  Skeleton
} from '@raystack/apsara/v1';
import { useSessions } from '../../../hooks/useSessions';
import { useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { RevokeSessionFinalConfirm } from './revoke-session-final-confirm';
import styles from './sessions.module.css';

export const RevokeSessionConfirm = () => {
  const navigate = useNavigate({ from: '/sessions/revoke' });
  const search = useSearch({ from: '/sessions/revoke' }) as { sessionId?: string };
  const { sessions, revokeSession, isRevokingSession } = useSessions();
  const { onLogout } = useRouteContext({ from: '__root__' }) as { onLogout?: () => void };
  const [isFinalConfirmOpen, setIsFinalConfirmOpen] = useState(false);

  const { mutate: logout } = useMutation(FrontierServiceQueries.authLogout, {
    onSuccess: () => {
      if (onLogout) {
        onLogout();
      }
    },
    onError: (error) => {
      console.error('Failed to logout:', error);
      // Fallback to regular session revocation
      if (search.sessionId) {
        revokeSession(search.sessionId);
        navigate({ to: '/sessions' });
      }
    },
  });

  // Find the session data based on sessionId from URL params
  const sessionData = useMemo(() => {
    if (!search.sessionId || sessions.length === 0) {
      return null;
    }
    
    const session = sessions.find(s => s.id === search.sessionId);
    if (!session) {
      console.error('Not found');
      return null;
    }
    
    return session;
  }, [search.sessionId, sessions]);

  const handleRevokeClick = () => {
    setIsFinalConfirmOpen(true);
  };

  const handleFinalConfirm = () => {
    if (!search.sessionId) return;
    
    if (sessionData?.isCurrent) {
      logout({});
      return;
    }
    
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
              <Flex justify="between" align="center" width="full">
                <Text size="regular">
                  {sessionData.browser} on {sessionData.operatingSystem}
                </Text>
                <Dialog.CloseButton data-test-id="frontier-sdk-close-revoke-session-dialog" />
              </Flex>
            </Dialog.Header>

            <Dialog.Body className={styles.revokeSessionConfirmBody}>
              <List className={styles.listRoot}>
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
              </List>
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
                  {sessionData?.isCurrent ? 'Logout' : 'Revoke'}
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
