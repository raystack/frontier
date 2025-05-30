import { Flex, Text } from '@raystack/apsara/v1';
import { PropsWithChildren } from 'react';
import styles from './layout.module.css';

interface LayoutProps {
  title: string;
}

export function Layout({ title, children }: PropsWithChildren<LayoutProps>) {
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex className={styles.header}>
        <Text size="large">{title}</Text>
      </Flex>
      <Flex direction="column" gap={9} className={styles.container}>
        {children}
      </Flex>
    </Flex>
  );
}
