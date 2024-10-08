import { Flex, Text } from '@raystack/apsara';
import { PropsWithChildren } from 'react';
import styles from './layout.module.css';

interface LayoutProps {
  title: string;
}

export function Layout({ title, children }: PropsWithChildren<LayoutProps>) {
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex className={styles.header}>
        <Text size={6}>{title}</Text>
      </Flex>
      <Flex direction="column" gap="large" className={styles.container}>
        {children}
      </Flex>
    </Flex>
  );
}
