'use client';

import {
  Text,
  Flex,
  Headline,
  Button,
} from '@raystack/apsara/v1';
import { Outlet, useNavigate } from '@tanstack/react-router';
import styles from './sessions.module.css';

export const SessionsPage = () => {
  const navigate = useNavigate({ from: '/sessions' });

  const handleRevoke = () => {
    navigate({ to: '/sessions/revoke' });
  };

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" className={styles.container}>
        <Flex direction="row" justify="between" align="center" className={styles.header}>
          <Flex direction="column" gap={2}>
            <Headline size="t1">
              Sessions
            </Headline>
            <Text size="regular" variant="secondary">
              Devices logged into this account.
            </Text>
          </Flex>
        </Flex>
        
        <Flex direction="column">
          <Flex justify="between" align="center" className={styles.sessionItem}>
            <Flex direction="column" gap={3}>
              <Text weight="medium" size="regular">Chrome on Mac OS x</Text>
              <Flex gap={2} align="center">
                <Text variant="tertiary" size="small">Bangalore</Text>
                <Text variant="tertiary" size="small">•</Text>
                <Text variant="success" size="small">Current session</Text>
              </Flex>
            </Flex>
            <Button variant="text" color="neutral" onClick={handleRevoke} data-test-id="frontier-sdk-revoke-session-button">Revoke</Button>
          </Flex>

          <Flex justify="between" align="center" className={styles.sessionItem}>
            <Flex direction="column" gap={3}>
              <Text weight="medium" size="regular">Chrome on Mac OS x</Text>
              <Flex gap={2} align="center">
                <Text variant="tertiary" size="small">Bangalore</Text>
                <Text variant="tertiary" size="small">•</Text>
                <Text variant="tertiary" size="small">Last active 10 minutes ago</Text>
              </Flex>
            </Flex>
            <Button variant="text" color="neutral" onClick={handleRevoke} data-test-id="frontier-sdk-revoke-session-button">Revoke</Button>
          </Flex>

        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
};