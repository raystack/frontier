import { Flex, Text } from '@raystack/apsara';
import { Outlet } from '@tanstack/react-router';
import { styles } from '../styles';

export default function Billing() {
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Billing</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}></Flex>
      </Flex>
      <Outlet />
    </Flex>
  );
}
