import { TrashIcon } from '@radix-ui/react-icons';
import { Avatar, Flex, Label, Text } from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import type { ColumnDef } from '@tanstack/react-table';
import Skeleton from 'react-loading-skeleton';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { getInitials } from '~/utils';

export const getColumns: (
  organizationId: string,
  canUpdateGroup?: boolean,
  memberRoles?: Record<string, Role[]>,
  isLoading?: boolean
) => ColumnDef<V1Beta1User, any>[] = (
  organizationId,
  canUpdateGroup = false,
  memberRoles = {},
  isLoading
) => [
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
          />
        )
  }
];

const MembersActions = ({
  member,
  organizationId,
  canUpdateGroup
}: {
  member: V1Beta1User;
  canUpdateGroup?: boolean;
  organizationId: string;
}) => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const { client } = useFrontier();
  const navigate = useNavigate({ from: '/teams/$teamId' });

  async function deleteMember() {
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
    <Flex align="center" justify="end" gap="large">
      <TrashIcon
        onClick={deleteMember}
        color="var(--foreground-danger)"
        style={{ cursor: 'pointer' }}
      />
    </Flex>
  ) : null;
};
