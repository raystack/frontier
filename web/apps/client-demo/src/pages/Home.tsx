import AuthContext from '@/contexts/auth';
import {
  Avatar,
  Button,
  DataTable,
  DropdownMenu,
  Flex,
  Navbar,
  Text,
  getAvatarColor,
  type DataTableColumnDef,
} from '@raystack/apsara';
import { useFrontier, useTerminology } from '@raystack/frontier/react';
import {
  useMutation,
  useQuery,
  FrontierServiceQueries,
  useQueryClient,
} from '@raystack/frontier/hooks';
import { useNavigate } from 'react-router-dom';
import { useContext, useEffect, useMemo, useCallback, useState, type MouseEvent } from 'react';
import { toast, IconButton, Separator } from '@raystack/apsara';
import { MagnifyingGlassIcon } from '@radix-ui/react-icons';

type OrgRow = {
  id: string;
  orgId: string;
  name: string;
  status: 'joined' | 'invited' | 'expired';
  timestamp: number;
  invitationId?: string;
};

const STATUS_LABELS: Record<OrgRow['status'], string> = {
  joined: 'Joined',
  invited: 'Invited',
  expired: 'Invite Expired',
};

function timeAgo(ts: number): string {
  if (!ts) return '-';
  const seconds = Math.floor((Date.now() - ts) / 1000);
  if (seconds < 60) return 'just now';
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  const months = Math.floor(days / 30);
  if (months < 12) return `${months}mo ago`;
  const years = Math.floor(months / 12);
  return `${years}y ago`;
}

function tsToMs(ts?: { seconds?: bigint; nanos?: number }): number {
  if (!ts?.seconds) return 0;
  return Number(ts.seconds) * 1000;
}

function getColumns(
  onAccept: (row: OrgRow) => void,
  onOpen: (row: OrgRow, e: MouseEvent) => void
): DataTableColumnDef<OrgRow, unknown>[] {
  return [
    {
      accessorKey: 'name',
      header: 'Organization',
      enableSorting: true,
      cell: ({ getValue }) => <Text weight="medium">{getValue() as string}</Text>,
    },
    {
      accessorKey: 'status',
      header: 'Status',
      enableSorting: true,
      enableGrouping: true,
      enableHiding: true,
      groupLabelsMap: STATUS_LABELS,
      cell: ({ row }) => <Text>{STATUS_LABELS[row.original.status]}</Text>,
    },
    {
      accessorKey: 'timestamp',
      header: 'Date',
      enableSorting: true,
      enableHiding: true,
      cell: ({ row }) => {
        const { status, timestamp } = row.original;
        const prefix = STATUS_LABELS[status];
        return (
          <Text variant="secondary">
            {timestamp ? `${prefix} ${timeAgo(timestamp)}` : '-'}
          </Text>
        );
      },
    },
    {
      accessorKey: 'id',
      header: '',
      cell: ({ row }) => {
        const { status } = row.original;
        if (status === 'invited') {
          return (
            <Button
              size="small"
              style={{ minWidth: 64 }}
              data-test-id={`[accept-invite-${row.original.invitationId}]`}
              onClick={(e: MouseEvent) => {
                e.stopPropagation();
                onAccept(row.original);
              }}
            >
              Join
            </Button>
          );
        }
        if (status === 'expired') {
          return (
            <Button
              size="small"
              style={{ minWidth: 64 }}
              disabled
            >
              Join
            </Button>
          );
        }
        if (status === 'joined') {
          return (
            <Button
              variant="outline"
              size="small"
              style={{ minWidth: 64 }}
              data-test-id={`[open-org-${row.original.orgId}]`}
              onClick={(e: MouseEvent) => {
                e.stopPropagation();
                onOpen(row.original, e);
              }}
            >
              Open
            </Button>
          );
        }
        return null;
      },
    },
  ];
}

export default function Home() {
  const { isAuthorized } = useContext(AuthContext);
  const { user, organizations } = useFrontier();
  const t = useTerminology();
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
      timestamp: tsToMs(org.createdAt),
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
        const status = isExpired ? ('expired' as const) : ('invited' as const);

        return {
          id: inv.id,
          orgId: inv.orgId,
          invitationId: inv.id,
          name: org?.title || org?.name || inv.orgId,
          status,
          timestamp: tsToMs(inv.createdAt),
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

  const handleOpen = useCallback(
    (row: OrgRow, e: MouseEvent) => {
      const path = `/organizations/${row.orgId}`;
      if (e.metaKey || e.ctrlKey) {
        window.open(path, '_blank');
      } else {
        navigate(path);
      }
    },
    [navigate],
  );

  const columns = useMemo(() => getColumns(handleAccept, handleOpen), [handleAccept, handleOpen]);

  async function logout() {
    try {
      await logoutMutation.mutateAsync({});
      window.location.reload();
    } catch (error) {
      console.error('Logout failed:', error);
    }
  }

  const [showSearch, setShowSearch] = useState(false);

  const avatarColor = getAvatarColor(user?.id || '');
  const userInitial = user?.title?.[0] || user?.email?.[0] || '?';

  return (
    <main style={{ height: '100vh', display: 'flex', flexDirection: 'column', margin: 0 }}>
      <DataTable
        columns={columns}
        data={rows}
        isLoading={isInvitationsLoading}
        defaultSort={{ name: 'status', order: 'asc' }}
      >
        <Navbar>
          <Navbar.Start>
            <Text size="large" weight="bold">
              {t.appName()}
            </Text>
          </Navbar.Start>
          <Navbar.End>
            <Flex align="center" gap="small">
              {showSearch ? (
                <DataTable.Search
                  autoFocus
                  showClearButton
                  size="small"
                  onBlur={(e) => {
                    if (!e.target.value) setShowSearch(false);
                  }}
                />
              ) : (
                <IconButton
                  size={3}
                  aria-label="Search"
                  onClick={() => setShowSearch(true)}
                >
                  <MagnifyingGlassIcon />
                </IconButton>
              )}
              <Separator orientation="vertical" size="small" />
              <DropdownMenu>
                <DropdownMenu.Trigger asChild>
                  <button
                    style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0 }}
                    data-test-id="[user-menu-trigger]"
                  >
                    <Avatar
                      src={user?.avatar}
                      fallback={userInitial}
                      color={avatarColor}
                      size={3}
                    />
                  </button>
                </DropdownMenu.Trigger>
                <DropdownMenu.Content>
                  <DropdownMenu.Item disabled>
                    {user?.email}
                  </DropdownMenu.Item>
                  <DropdownMenu.Item
                    onClick={logout}
                    data-test-id="[logout-button]"
                  >
                    Logout
                  </DropdownMenu.Item>
                </DropdownMenu.Content>
              </DropdownMenu>
            </Flex>
          </Navbar.End>
        </Navbar>
        <Flex direction="column" style={{ flex: 1, padding: 'var(--rs-space-5)' }}>
          <DataTable.Toolbar />
          <DataTable.Content />
        </Flex>
      </DataTable>
    </main>
  );
}
