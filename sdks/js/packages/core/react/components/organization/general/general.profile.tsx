'use client';

import { Avatar, Flex, Text } from '@raystack/apsara';

// @ts-ignore
import styles from './general.module.css';

export const GeneralProfile = () => {
  return (
    <Flex direction="column" gap="small">
      <Avatar
        alt="Colm Tuite"
        shape="circle"
        fallback="CT"
        imageProps={{ width: '80px', height: '80px' }}
      />
      <Text size={4} className={styles.profileDescription}>
        Pick a logo for your organisation. Max size: 5 Mb
      </Text>
    </Flex>
  );
};
