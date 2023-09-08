'use client';

import { Avatar, Flex } from '@raystack/apsara';
import Skeleton from 'react-loading-skeleton';

// @ts-ignore
import { V1Beta1Organization } from '~/src';
import { getInitials } from '~/utils';
// import styles from './general.module.css';

interface GeneralProfileProps {
  organization?: V1Beta1Organization;
  isLoading?: boolean;
}
export const GeneralProfile = ({
  organization,
  isLoading
}: GeneralProfileProps) => {
  return (
    <Flex direction="column" gap="small">
      {isLoading ? (
        <Skeleton style={{ width: '80px', height: '80px' }} circle />
      ) : (
        <Avatar
          alt="Organization profile"
          shape="circle"
          fallback={getInitials(organization?.name)}
          imageProps={{ width: '80px', height: '80px' }}
        />
      )}
      {/* <Text size={4} className={styles.profileDescription}>
        Pick a logo for your organisation. Max size: 5 Mb
      </Text> */}
    </Flex>
  );
};
