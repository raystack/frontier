'use client';

import { Flex } from '@raystack/apsara';
import { Text } from '@raystack/apsara/v1';
import { styles } from '../styles';
import { UpdateProfile } from './update';

export function UserSetting() {
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size="large">Profile</Text>
      </Flex>
      <UpdateProfile />
    </Flex>
  );
}
