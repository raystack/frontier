'use client';

import {
  Text,
  Flex,
  Headline,
} from '@raystack/apsara/v1';
import { Outlet } from '@tanstack/react-router';
import styles from './sessions.module.css';

export default function SessionsPage() {

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex direction="column" gap={9} className={styles.container}>
        <Flex direction="row" justify="between" align="center">
        <Flex direction="column" gap={3}>
          <Headline size="t1">
            Session
          </Headline>
          <Text size="regular" variant="secondary">
            Devices logged into this account.
          </Text>
        </Flex>
      </Flex>
        {/* Sessions list will go here */}
      </Flex>
      <Outlet />
    </Flex>
  );
}