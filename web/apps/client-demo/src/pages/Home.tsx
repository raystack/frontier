import AuthContext from '@/contexts/auth';
import {
  Avatar,
  Button,
  DataTable,
  Flex,
  Navbar,
  Text,
  getAvatarColor,
  type DataTableColumnDef,
} from '@raystack/apsara';
import { useFrontier } from '@raystack/frontier/react';
import {
  useMutation,
  useQuery,
  FrontierServiceQueries,
  useQueryClient,
} from '@raystack/frontier/hooks';
import { useNavigate } from 'react-router-dom';
import { useContext, useEffect, useMemo, useCallback } from 'react';
import { toast } from '@raystack/apsara';

type OrgRow = {
  id: string;
  orgId: string;
  name: string;
  status: 'joined' | 'invited' | 'expired';
  date: string;
  invitationId?: string;
};

const STATUS_LABELS: Record<OrgRow['status'], string> = {
  joined: 'Joined',
  invited: 'Invited',
  expired: 'Invite Expired',
};

function getColumns(
  onAccept: (row: OrgRow) => void
): DataTableColumnDef<OrgRow, unknown>[] {
  return [
    {
      accessorKey: 'name',
      header: 'Organization',
      enableColumnFilter: true,
      cell: ({ getValue }) => <Text weight="medium">{getValue() as string}</Text>,
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => {
        const status = row.original.status;
        if (status === 'invited') {
          return (
            <Flex gap="small" align="center">
              <Text>{STATUS_LABELS[status]}</Text>
              <Button
                size="small"
                data-test-id={`[accept-invite-${row.original.invitationId}]`}
                onClick={(e) => {
                  e.stopPropagation();
                  onAccept(row.original);
                }}
              >
                Accept
              </Button>
            </Flex>
          );
        }
        return <Text>{STATUS_LABELS[status]}</Text>;
      },
      filterType: 'select',
      filterOptions: Object.entries(STATUS_LABELS).map(([value, label]) => ({
        value,
        label,
      })),
      enableColumnFilter: true,
    },
    {
      accessorKey: 'date',
      header: 'Date',
      cell: ({ getValue }) => <Text variant="secondary">{getValue() as string}</Text>,
    },
  ];
}

function formatDate(ts?: { seconds?: bigint; nanos?: number }): string {
  if (!ts?.seconds) return '-';
  return new Date(Number(ts.seconds) * 1000).toLocaleDateString();
}

export default function Home() {
  const { isAuthorized } = useContext(AuthContext);
  const { user, organizations } = useFrontier();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const logoutMutation = useMutation(FrontierServiceQueries.authLogout);

  const {
    data: invitationsData,
    isLoading: isInvitationsLoading,
  } = useQuery(FrontierServiceQueries.listCurrentUserInvitations, {}, {
    enabled: isAuthorized,
  });

  const { mutateAsync: acceptInvitation } = useMutation(
    FrontierServiceQueries.acceptOrganizationInvitation,
  );

  useEffect(() => {
    if (!isAuthorized) {
      navigate('/login');
    }
  }, [isAuthorized, navigate]);

  const invitations = invitationsData?.invitations ?? [];
  const inviteOrgs = invitationsData?.orgs ?? [];

  const inviteOrgMap = useMemo(
    () =>
      inviteOrgs.reduce(
        (acc, org) => {
          acc[org.id] = org;
          return acc;
        },
        {} as Record<string, { id: string; title: string; name: string }>,
      ),
    [inviteOrgs],
  );

  const rows: OrgRow[] = useMemo(() => {
    const now = Date.now();

    const joinedRows: OrgRow[] = organizations.map((org) => ({
      id: org.id,
      orgId: org.id,
      name: org.title || org.name || org.id,
      status: 'joined' as const,
      date: formatDate(org.createdAt),
    }));

    const joinedOrgIds = new Set(organizations.map((o) => o.id));

    const inviteRows: OrgRow[] = invitations
      .filter((inv) => !joinedOrgIds.has(inv.orgId))
      .map((inv) => {
        const org = inviteOrgMap[inv.orgId];
        const expiresMs = inv.expiresAt?.seconds
          ? Number(inv.expiresAt.seconds) * 1000
          : 0;
        const isExpired = expiresMs > 0 && expiresMs < now;

        return {
          id: inv.id,
          orgId: inv.orgId,
          invitationId: inv.id,
          name: org?.title || org?.name || inv.orgId,
          status: isExpired ? ('expired' as const) : ('invited' as const),
          date: formatDate(inv.createdAt),
        };
      });

    return [...inviteRows, ...joinedRows];
  }, [organizations, invitations, inviteOrgMap]);

  const handleAccept = useCallback(
    async (row: OrgRow) => {
      try {
        await acceptInvitation({ id: row.invitationId!, orgId: row.orgId });
        toast.success('Invitation accepted');
        queryClient.invalidateQueries();
      } catch (err) {
        const message = err instanceof Error ? err.message : 'Something went wrong';
        toast.error(`Failed to accept: ${message}`);
      }
    },
    [acceptInvitation, queryClient],
  );

  const columns = useMemo(() => getColumns(handleAccept), [handleAccept]);

  async function logout() {
    try {
      await logoutMutation.mutateAsync({});
      window.location.reload();
    } catch (error) {
      console.error('Logout failed:', error);
    }
  }

  const avatarColor = getAvatarColor(user?.id || '');
  const userInitial = user?.title?.[0] || user?.email?.[0] || '?';

  return (
    <main style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Navbar>
        <Navbar.Start>
          <Text size="large" weight="bold">
            Frontier
          </Text>
        </Navbar.Start>
        <Navbar.End>
          <Flex align="center" gap="small">
            <Avatar
              src={user?.avatar}
              fallback={userInitial}
              color={avatarColor}
              size={2}
            />
            <Text size="small">{user?.title || user?.email}</Text>
            <Button
              variant="outline"
              color="neutral"
              size="small"
              data-test-id="[logout-button]"
              onClick={logout}
            >
              Logout
            </Button>
          </Flex>
        </Navbar.End>
      </Navbar>
      <Flex direction="column" style={{ flex: 1, padding: 'var(--rs-space-5)' }}>
        <DataTable
          columns={columns}
          data={rows}
          isLoading={isInvitationsLoading}
          defaultSort={{ name: 'status', order: 'asc' }}
        >
          <Flex direction="column" style={{ width: '100%' }}>
            <DataTable.Toolbar />
            <DataTable.Content />
          </Flex>
        </DataTable>
      </Flex>
    </main>
  );
}
