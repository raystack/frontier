'use client';

import { Flex, Text } from '@raystack/apsara';
import { styles } from '../styles';
import { UpdateProfile } from './update';

export function UserSetting() {
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Profile</Text>
      </Flex>
      <UpdateProfile />
    </Flex>
  );
}
