'use client';

import { Flex, Separator, Text } from '@raystack/apsara';
import { styles } from '../styles';
import { GeneralProfile } from './avatar';
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
