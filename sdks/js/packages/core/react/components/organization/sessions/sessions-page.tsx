'use client';

import {
  Text,
  Flex,
  Headline,
  Button,
  Skeleton,
} from '@raystack/apsara/v1';
import { Outlet, useNavigate } from '@tanstack/react-router';
import { useSessions } from '../../../hooks/useSessions';
import styles from './sessions.module.css';

export const SessionsPage = () => {
  const navigate = useNavigate({ from: '/sessions' });
  const { sessions, isLoading, error } = useSessions();


  const handleRevoke = (sessionId: string) => {
    navigate({ to: '/sessions/revoke', search: { sessionId } });
  };

  const renderSessionsHeader = () => (
    <Flex direction="column" gap={2}>
      <Headline size="t1">Sessions</Headline>
      <Text size="regular" variant="secondary">
        Devices logged into this account.
      </Text>
    </Flex>
  );

  if (isLoading) {
    return (
      <Flex direction="column" width="full">
        <Flex direction="column" className={styles.container}>
          <Flex direction="row" justify="between" align="center" className={styles.header}>
            {renderSessionsHeader()}
          </Flex>
          <Flex direction="column" className={styles.sessionsList}>
            <Skeleton 
              height="60px" 
              containerStyle={{ padding: '1rem 0' }}
              count={3}
            />
          </Flex>
        </Flex>
      </Flex>
    );
  }

  if (error) {
    return (
      <Flex direction="column" width="full">
        <Flex direction="column" className={styles.container}>
          <Flex direction="row" justify="between" align="center" className={styles.header}>
            {renderSessionsHeader()}
          </Flex>
          <Flex justify="center" align="center" style={{ padding: '2rem' }}>
            <Text variant="danger" size="regular">
              {error}
            </Text>
          </Flex>
        </Flex>
      </Flex>
    );
  }

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" className={styles.container}>
        <Flex direction="row" justify="between" align="center" className={styles.header}>
          {renderSessionsHeader()}
        </Flex>
        
        <Flex direction="column" className={styles.sessionsList}>
          {sessions.length === 0 ? (
            <Flex justify="center" align="center" style={{ padding: '2rem' }}>
              <Text variant="tertiary" size="regular">
                No active sessions found.
              </Text>
            </Flex>
          ) : (
            sessions.map((session) => (
              <Flex key={session.id} justify="between" align="center" className={styles.sessionItem}>
                <Flex direction="column" gap={3}>
                  <Text weight="medium" size="regular">
                    {session.browser} on {session.operatingSystem}
                  </Text>
                  <Flex gap={2} align="center">
                    <Text variant="tertiary" size="small">{session.location}</Text>
                    <Text variant="tertiary" size="small">•</Text>
                    {session.isCurrent ? (
                      <Text variant="success" size="small">Current session</Text>
                    ) : (
                      <Text variant="tertiary" size="small">Last active {session.lastActive}</Text>
                    )}
                  </Flex>
                </Flex>
                <Button 
                  variant="text" 
                  color="neutral" 
                  onClick={() => handleRevoke(session.id)} 
                  data-test-id="frontier-sdk-revoke-session-button"
                >
                  {session.isCurrent ? 'Logout' : 'Revoke'}
                </Button>
              </Flex>
            ))
          )}
        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
};