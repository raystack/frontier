'use client';

import { Flex, Skeleton } from '@raystack/apsara';
import styles from './sessions.module.css';

export const SessionSkeleton = ({ count = 3 }: { count?: number }) => {
  return (
    <>
      {Array.from({ length: count }).map((_, index) => (
        <Flex key={index} justify="between" align="center" className={styles.sessionItem}>
          <Flex direction="column" gap={3}>
            <Skeleton height="18px" width="200px" />
            <Flex gap={2} align="center">
              <Skeleton height="16px" width="120px" />
            </Flex>
          </Flex>
          <Skeleton height="32px" width="64px" borderRadius="var(--rs-radius-2)" />
        </Flex>
      ))}
    </>
  );
};

