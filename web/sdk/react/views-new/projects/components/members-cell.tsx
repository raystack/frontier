'use client';

import {
  Avatar,
  Flex,
  Skeleton,
  Text,
  getAvatarColor
} from '@raystack/apsara-v1';
import { useQuery } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  ListProjectUsersRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { getInitials } from '~/utils';
import styles from './members-cell.module.css';

const MAX_AVATARS = 4;

export interface MembersCellProps {
  projectId: string;
}

export function MembersCell({ projectId }: MembersCellProps) {
  const { data: users = [], isLoading } = useQuery(
    FrontierServiceQueries.listProjectUsers,
    create(ListProjectUsersRequestSchema, { id: projectId }),
    {
      enabled: !!projectId,
      select: (d) => d?.users ?? []
    }
  );

  if (isLoading) return <Skeleton height="24px" width="120px" />;

  if (!users.length) return null;

  return (
    <Flex gap={3} align="center" style={{ maxWidth: '120px' }}>
      <Flex align="center">
        {users.slice(0, MAX_AVATARS).map(user => (
          <Avatar
            key={user.id}
            src={user.avatar}
            fallback={getInitials(user.title || user.email || user.id)}
            radius="full"
            size={3}
            color={getAvatarColor(user.id)}
            className={styles.avatar}
          />
        ))}
      </Flex>
      {users.length > MAX_AVATARS && <Text size="small" variant="secondary">+{users.length - MAX_AVATARS}</Text>}
    </Flex>
  );
}
