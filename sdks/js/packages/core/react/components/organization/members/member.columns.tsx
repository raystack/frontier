import { TrashIcon } from '@radix-ui/react-icons';
import { Avatar, Flex, Label, Text } from '@raystack/apsara';
import type { ColumnDef } from '@tanstack/react-table';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1User } from '~/src';
import { getInitials } from '~/utils';
import Skeleton from 'react-loading-skeleton';

export const getColumns: (
  id: string,
  isLoading?: boolean
) => ColumnDef<V1Beta1User, any>[] = (organizationId, isLoading) => [
  {
    header: '',
    accessorKey: 'profile_picture',
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
              fallback={getInitials(row.original?.title)}
              // @ts-ignore
              style={{ marginRight: 'var(--mr-12)', zIndex: -1 }}
            />
          );
        }
  },
  {
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
    accessorKey: 'email',
    meta: {},
    cell: isLoading
      ? () => <Skeleton />
      : ({ row, getValue }) => {
          return (
            // @ts-ignore
            <Text>{getValue() || row.original?.user_id}</Text>
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
          />
        )
  }
];

const MembersActions = ({
  member,
  organizationId
}: {
  member: V1Beta1User;
  organizationId: string;
}) => {
  const { client } = useFrontier();
  const navigate = useNavigate({ from: '/members' });

  async function deleteMember() {
    try {
      // @ts-ignore
      if (member?.invited) {
        await client?.frontierServiceDeleteOrganizationInvitation(
          // @ts-ignore
          member.org_id,
          member?.id as string
        );
      } else {
        await client?.frontierServiceRemoveOrganizationUser(
          organizationId,
          member?.id as string
        );
      }
      navigate({ to: '/members' });
      toast.success('Member deleted');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <Flex align="center" justify="end" gap="large">
      <TrashIcon
        onClick={deleteMember}
        color="var(--foreground-danger)"
        style={{ cursor: 'pointer' }}
      />
    </Flex>
  );
};
