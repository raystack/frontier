'use client';

import { Avatar, Flex, Text } from '@raystack/apsara';
import Skeleton from 'react-loading-skeleton';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { getInitials } from '~/utils';

export const GeneralProfile = () => {
  const { user, isUserLoading: isLoading } = useFrontier();
  return (
    <Flex direction="column" gap="small">
      {isLoading ? (
        <Skeleton style={{ width: '80px', height: '80px' }} />
      ) : (
        <Avatar
          alt="User profile"
          shape="square"
          fallback={getInitials(user?.title || user?.email)}
          imageProps={{ width: '80px', height: '80px' }}
        />
      )}

      {/* <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
        Pick a profile picture for your avatar. Max size: 5 Mb
      </Text> */}
    </Flex>
  );
};
