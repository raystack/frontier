'use client';

import { Avatar, Flex, Text } from '@raystack/apsara';

// @ts-ignore
import { V1Beta1Organization } from '~/src';
import { getInitials } from '~/utils';
import styles from './general.module.css';

interface GeneralProfileProps {
  organization?: V1Beta1Organization;
}
export const GeneralProfile = ({ organization }: GeneralProfileProps) => {
  return (
    <Flex direction="column" gap="small">
      <Avatar
        alt="Organization profile"
        shape="circle"
        fallback={getInitials(organization?.name)}
        imageProps={{ width: '80px', height: '80px' }}
      />
      <Text size={4} className={styles.profileDescription}>
        Pick a logo for your organisation. Max size: 5 Mb
      </Text>
    </Flex>
  );
};
