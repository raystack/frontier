import {
  DotsHorizontalIcon,
  TrashIcon,
  UpdateIcon
} from '@radix-ui/react-icons';
import { Avatar, DropdownMenu, Flex, Label, Text } from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import Skeleton from 'react-loading-skeleton';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { getInitials } from '~/utils';
import styles from '../../organization.module.css';
import { useMemo } from 'react';

interface getColumnsOptions {
  organizationId: string;
  canUpdateGroup?: boolean;
  memberRoles?: Record<string, Role[]>;
  isLoading?: boolean;
  membersCount: number;
}

export const getColumns: (
  options: getColumnsOptions
) => ColumnDef<V1Beta1User, any>[] = ({
  organizationId,
  canUpdateGroup = false,
  memberRoles = {},
  isLoading,
  membersCount
}) => [
  {
    header: '',
    accessorKey: 'image',
    size: 44,
    meta: {
      style: {
        width: '30px',
        padding: 0
      }
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          return (
            <Avatar
              src={getValue()}
              fallback={getInitials(row.original?.title || row.original?.email)}
              // @ts-ignore
              style={{ marginRight: 'var(--mr-12)' }}
            />
          );
        }
  },
  {
    header: 'Title',
    accessorKey: 'title',
    meta: {
      style: {
        paddingLeft: 0
      }
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          return (
            <Flex direction="column" gap="extra-small">
              <Label style={{ fontWeight: '$500' }}>{getValue()}</Label>
              <Text>{row.original.email}</Text>
            </Flex>
          );
        }
  },
  {
    header: 'Roles',
    accessorKey: 'email',
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          return (
            (row.original?.id &&
              memberRoles[row.original?.id] &&
              memberRoles[row.original?.id]
                .map((r: any) => r.title || r.name)
                .join(', ')) ??
            'Inherited role'
          );
        }
  },
  {
    header: '',
    accessorKey: 'id',
    meta: {
      style: {
        textAlign: 'end'
      }
    },
    cell: isLoading
      ? () => <Skeleton />
      : ({ row }) => (
          <MembersActions
            member={row.original as V1Beta1User}
            organizationId={organizationId}
            canUpdateGroup={canUpdateGroup}
            membersCount={memberRoles}
          />
        )
  }
];

const MembersActions = ({
  member,
  organizationId,
  canUpdateGroup,
  membersCount
}: {
  member: V1Beta1User;
  canUpdateGroup?: boolean;
  organizationId: string;
  membersCount: number;
}) => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const { client } = useFrontier();
  const navigate = useNavigate({ from: '/teams/$teamId' });

  // TODO: add check if only admin can remove themself.
  const canRemoveSelf = useMemo(() => membersCount > 1, [membersCount]);
  // TODO: add check for other member remove
  const canRemove = canRemoveSelf;

  async function deleteMember() {
    if (!canRemove) return;
    try {
      await client?.frontierServiceRemoveGroupUser(
        organizationId,
        teamId as string,
        member?.id as string
      );
      navigate({
        to: '/teams/$teamId',
        params: {
          teamId
        }
      });
      toast.success('Member deleted');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return canUpdateGroup ? (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild style={{ cursor: 'pointer' }}>
        <DotsHorizontalIcon />
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Group>
          <DropdownMenu.Item style={{ padding: 0 }} disabled={!canRemoveSelf}>
            <div onClick={deleteMember} className={styles.dropdownActionItem}>
              <TrashIcon />
              Remove from team
            </div>
          </DropdownMenu.Item>
        </DropdownMenu.Group>
      </DropdownMenu.Content>
    </DropdownMenu>
  ) : null;
};
