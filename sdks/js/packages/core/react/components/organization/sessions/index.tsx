'use client';

import {
  Text,
  Flex,
  Headline,
  Button,
} from '@raystack/apsara/v1';
import { Outlet } from '@tanstack/react-router';
import styles from './sessions.module.css';

export default function SessionsPage() {
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" className={styles.container}>
        <Flex direction="row" justify="between" align="center" className={styles.header}>
          <Flex direction="column" gap={2}>
            <Headline size="t1">
              Session
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
                <Text variant="tertiary" size="micro">Bangalore</Text>
                <Text variant="tertiary" size="micro">•</Text>
                <Text variant="success" size="micro">Current session</Text>
              </Flex>
            </Flex>
            <Button variant="text" color="neutral">Revoke</Button>
          </Flex>

          <Flex justify="between" align="center" className={styles.sessionItem}>
            <Flex direction="column" gap={3}>
              <Text weight="medium" size="regular">Chrome on Mac OS x</Text>
              <Flex gap={2} align="center">
                <Text variant="tertiary" size="micro">Bangalore</Text>
                <Text variant="tertiary" size="micro">•</Text>
                <Text variant="tertiary" size="micro">Last active 10 minutes ago</Text>
              </Flex>
            </Flex>
            <Button variant="text" color="neutral">Revoke</Button>
          </Flex>

        </Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}