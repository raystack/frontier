'use client';

import {
  Avatar,
  AvatarGroup,
  Skeleton,
  getAvatarColor
} from '@raystack/apsara-v1';
import { useQuery } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  ListProjectUsersRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { getInitials } from '~/utils';

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
    <AvatarGroup max={MAX_AVATARS}>
      {users.map(user => (
        <Avatar
          key={user.id}
          src={user.avatar}
          fallback={getInitials(user.title || user.email || user.id)}
          radius="full"
          size={3}
          color={getAvatarColor(user.id)}
        />
      ))}
    </AvatarGroup>
  );
}
