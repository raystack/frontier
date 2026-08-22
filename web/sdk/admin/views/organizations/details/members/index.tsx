import { AlertDialog, Button, DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import { PageTitle } from "~/admin/components/PageTitle";
import styles from "./members.module.css";
import { useContext, useEffect, useMemo, useRef, useState } from "react";
import { getColumns } from "./columns";
import type {
  Invitation,
  SearchOrganizationUsersResponse_OrganizationUser,
} from "@raystack/proton/frontier";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
  ListOrganizationInvitationsRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import {
  useInfiniteQuery,
  useQuery,
  createConnectQueryKey,
  useTransport
} from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { UsersIcon } from '../../../../assets/icons/UsersIcon';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { OrganizationContext } from '../contexts/organization-context';
import { UpdateRole, type UpdateRolePayload } from './update-role';
import { RemoveMember } from './remove-member';
import {
  getConnectNextPageParam,
  DEFAULT_PAGE_SIZE
} from '~/utils/connect-pagination';
import { transformDataTableQueryToRQLRequest } from '~/utils/transform-query';
import { useDebouncedValue } from '~hooks';
import { useTerminology } from "~/admin/hooks/useTerminology";
import { InvitedMembersDialog } from './invited-members-dialog';

const updateRoleDialogHandle = AlertDialog.createHandle<UpdateRolePayload>();

// Stable ref: a fresh [] each render would remount the invites table.
const NO_INVITATIONS: Invitation[] = [];

const DEFAULT_SORT: DataTableSort = { name: 'orgJoinedAt', order: 'desc' };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
  // Must match DataTable's mount emit, or it refetches.
  sort: [DEFAULT_SORT],
};
const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    orgJoinedAt: "org_joined_at",
    roleIds: "role_ids",
    createdAt: "created_at",
    updatedAt: "updated_at",
  },
};

const NoMembers = () => {
  const t = useTerminology();
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading={`No ${t.member({ case: "capital" })} found`}
      subHeading="No results found matching the selected filters. Try adjusting or resetting them to view more results."
      icon={<UsersIcon />}
    />
  );
};

const ZeroState = () => {
  const t = useTerminology();
  return (
    <div className={styles["zero-state-container"]}>
      <EmptyState
        variant="empty2"
        icon={<UsersIcon />}
        heading={t.member({ plural: true, case: "capital" })}
        subHeading={`${t.member({ plural: true, case: "capital" })} are ${t.user({ plural: true, case: "lower" })} who belong to this ${t.organization({ case: "lower" })} and can access its resources.`}
      />
    </div>
  );
};

const ErrorState = () => {
  const t = useTerminology();
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading={`Error Loading ${t.member({ plural: true, case: "capital" })}`}
      subHeading={`Something went wrong while loading ${t.organization({ case: "lower" })} ${t.member({ plural: true, case: "lower" })}. Please try refreshing the page.`}
      icon={<ExclamationTriangleIcon />}
    />
  );
};

export function OrganizationMembersView() {
  const t = useTerminology();
  const { roles = [], organization, search } = useContext(OrganizationContext);
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;
  const queryClient = useQueryClient();
  const transport = useTransport();

  const organizationId = organization?.id || "";

  const [removeMemberConfig, setRemoveMemberConfig] = useState<{
    isOpen: boolean;
    user: SearchOrganizationUsersResponse_OrganizationUser | null;
  }>({ isOpen: false, user: null });

  const [isInvitesDialogOpen, setIsInvitesDialogOpen] = useState(false);

  // Not in the dialog: the toolbar needs the count before it mounts.
  const {
    data: invitations = NO_INVITATIONS,
    isLoading: isInvitationsLoading,
    error: invitationsError,
  } = useQuery(
    FrontierServiceQueries.listOrganizationInvitations,
    create(ListOrganizationInvitationsRequestSchema, {
      orgId: organizationId,
    }),
    {
      enabled: !!organizationId,
      select: data => data?.invitations || NO_INVITATIONS,
    },
  );

  const title = `${t.member({ plural: true, case: "capital" })} | ${organization?.title} | ${t.organization({ plural: true, case: "capital" })}`;

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const computedQuery = useMemo(() => {
    const tempQuery = transformDataTableQueryToRQLRequest(tableQuery, TRANSFORM_OPTIONS);
    return {
      ...tempQuery,
      search: searchQuery || "",
    };
  }, [tableQuery, searchQuery]);

  const query = useDebouncedValue(computedQuery, 200);

  const {
    data: infiniteData,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    isError,
  } = useInfiniteQuery(
    AdminServiceQueries.searchOrganizationUsers,
    { id: organizationId, query: query },
    {
      enabled: !!organizationId,
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(lastPage, { query: query }, "orgUsers"),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const data = infiniteData?.pages?.flatMap(page => page.orgUsers) || [];
  // The backend doesn't send total_count, so rely on the loaded rows. This is
  // only used to prevent removing the last remaining member.
  const memberCount = data.length;
  const loading = (isLoading || isFetchingNextPage) && !isError;

  /*
   * DataTable seeds its query once at mount, so it never sees the org-context
   * search. Hence picking the state here instead of via the zeroState prop.
   */
  const hasActiveQuery = Boolean(
    searchQuery?.trim() || tableQuery.filters?.length,
  );
  const showZeroState =
    !isLoading && !isError && !hasActiveQuery && data.length === 0;

  // DataTable.Toolbar's own rule: hidden in the zero state.
  const showToolbar = data.length > 0 || Boolean(tableQuery.filters?.length);
  /*
   * - an org always has at least one member, so the toolbar-less branch below
   *   is a safeguard, not a state to expect
   * - the trigger is the count: a failed fetch leaves nothing to label, so it
   *   stays hidden and only logs
   */
  const showInvitesBtn = invitations.length > 0;

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery(newQuery);
  };

  // isFetchingNextPage lags a render; the ref doesn't.
  const isLoadingMoreRef = useRef(false);
  const fetchMore = async () => {
    if (!hasNextPage || isFetchingNextPage || isError || isLoadingMoreRef.current) return;
    isLoadingMoreRef.current = true;
    try {
      await fetchNextPage();
    } finally {
      isLoadingMoreRef.current = false;
    }
  };

  useEffect(() => {
    if (invitationsError) {
      console.error(
        "Failed to fetch organization invitations:",
        invitationsError,
      );
    }
  }, [invitationsError]);

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  function openRemoveMemberDialog(
    user: SearchOrganizationUsersResponse_OrganizationUser,
  ) {
    setRemoveMemberConfig({ isOpen: true, user });
  }

  function closeRemoveMemberDialog() {
    setRemoveMemberConfig({ isOpen: false, user: null });
  }

  const columns = getColumns({
    roles,
    memberCount,
    updateRoleHandle: updateRoleDialogHandle,
    handleRemoveMemberAction: openRemoveMemberDialog,
  });

  async function invalidateMembersQuery() {
    // Keys match partially: {} would hit every org; omitting query is deliberate.
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: AdminServiceQueries.searchOrganizationUsers,
        transport,
        input: { id: organizationId },
        cardinality: "infinite",
      }),
    });
  }

  async function updateMember() {
    // Invalidate and refetch the query
    await invalidateMembersQuery();
  }

  async function removeMember(
    userToRemove: SearchOrganizationUsersResponse_OrganizationUser,
  ) {
    setRemoveMemberConfig({ isOpen: false, user: null });
    // Invalidate and refetch the query
    await invalidateMembersQuery();
  }

  return (
    <>
      <UpdateRole
        handle={updateRoleDialogHandle}
        organizationId={organizationId}
        onRoleUpdate={updateMember}
      />

      {removeMemberConfig.isOpen && removeMemberConfig.user ? (
        <RemoveMember
          organizationId={organizationId}
          user={removeMemberConfig.user}
          onRemove={removeMember}
          onClose={closeRemoveMemberDialog}
        />
      ) : null}

      {isInvitesDialogOpen ? (
        <InvitedMembersDialog
          organizationId={organizationId}
          invitations={invitations}
          isLoading={isInvitationsLoading}
          onClose={() => setIsInvitesDialogOpen(false)}
        />
      ) : null}
      <Flex justify="center" className={styles["container"]}>
        <PageTitle title={title} />
        <DataTable
          columns={columns}
          data={data}
          isLoading={loading}
          defaultSort={DEFAULT_SORT}
          mode="server"
          onTableQueryChange={onTableQueryChange}
          onLoadMore={fetchMore}
          query={tableQuery}>
          <Flex direction="column" style={{ width: "100%" }}>
            {/* DataTable.Toolbar takes no children, so the row is rebuilt from
                its parts to seat the invites trigger left of Display. */}
            {(showToolbar || showInvitesBtn) && (
              <Flex
                justify={showToolbar ? "between" : "end"}
                align="start"
                gap={3}
                className={styles["toolbar"]}>
                {showToolbar && <DataTable.Filters />}
                <Flex align="center" gap={3}>
                  {showInvitesBtn && (
                    <Button
                      variant="text"
                      color="neutral"
                      size="small"
                      onClick={() => setIsInvitesDialogOpen(true)}
                      data-test-id="admin-org-members-invites">
                      {/* Expired invites come back too, so "Pending" would
                          overstate the count — Status labels each row. */}
                      {`${invitations.length} Invite${invitations.length === 1 ? "" : "s"}`}
                    </Button>
                  )}
                  {showToolbar && <DataTable.DisplayControls />}
                </Flex>
              </Flex>
            )}
            <DataTable.Content
              emptyState={showZeroState ? <ZeroState /> : isError ? <ErrorState /> : <NoMembers />}
              classNames={{
                table: styles["table"],
                root: styles["table-wrapper"],
                header: styles["table-header"],
              }}
            />
          </Flex>
        </DataTable>
      </Flex>
    </>
  );
}
