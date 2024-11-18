import { Flex, Text } from '@raystack/apsara';
import { styles } from '../styles';

export default function ApiKeys() {
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>API Keys</Text>
      </Flex>
    </Flex>
  );
}
