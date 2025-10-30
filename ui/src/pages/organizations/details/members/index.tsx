import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import PageTitle from "~/components/page-title";
import styles from "./members.module.css";
import { useContext, useEffect, useState } from "react";
import { getColumns } from "./columns";
import type { SearchOrganizationUsersResponse_OrganizationUser } from "@raystack/proton/frontier";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import {
  useInfiniteQuery,
  createConnectQueryKey,
  useTransport,
} from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import UserIcon from "~/assets/icons/users.svg?react";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { OrganizationContext } from "../contexts/organization-context";
import { AssignRole } from "~/components/assign-role";
import { RemoveMember } from "./remove-member";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import {
  getConnectNextPageParam,
  DEFAULT_PAGE_SIZE,
} from "~/utils/connect-pagination";
import { useDebouncedState } from "~/hooks/useDebouncedState";

const DEFAULT_SORT: DataTableSort = { name: "org_joined_at", order: "desc" };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};

const NoMembers = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Member found"
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<UserIcon />}
    />
  );
};

const ErrorState = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="Error Loading Members"
      subHeading="Something went wrong while loading organization members. Please try refreshing the page."
      icon={<ExclamationTriangleIcon />}
    />
  );
};

export function OrganizationMembersPage() {
  const { roles = [], organization, search } = useContext(OrganizationContext);
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;
  const queryClient = useQueryClient();
  const transport = useTransport();

  const organizationId = organization?.id || "";

  const [assignRoleConfig, setAssignRoleConfig] = useState<{
    isOpen: boolean;
    user: SearchOrganizationUsersResponse_OrganizationUser | null;
  }>({ isOpen: false, user: null });
  const [removeMemberConfig, setRemoveMemberConfig] = useState<{
    isOpen: boolean;
    user: SearchOrganizationUsersResponse_OrganizationUser | null;
  }>({ isOpen: false, user: null });

  const title = `Members | ${organization?.title} | Organizations`;

  const [tableQuery, setTableQuery] = useDebouncedState<DataTableQuery>(
    INITIAL_QUERY,
    200,
  );

  // Transform the DataTableQuery to RQLRequest format
  const query = transformDataTableQueryToRQLRequest(tableQuery, {
    fieldNameMapping: {
      orgJoinedAt: "org_joined_at",
      roleIds: "role_ids",
      createdAt: "created_at",
      updatedAt: "updated_at",
    },
  });

  // Add search to the query if present
  if (searchQuery) {
    query.search = searchQuery;
  }

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
  const loading = (isLoading || isFetchingNextPage) && !isError;

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery(newQuery);
  };

  const fetchMore = async () => {
    if (hasNextPage && !isFetchingNextPage && !isError) {
      await fetchNextPage();
    }
  };

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  function openAssignRoleDialog(
    user: SearchOrganizationUsersResponse_OrganizationUser,
  ) {
    setAssignRoleConfig({ isOpen: true, user });
  }

  function closeAssignRoleDialog() {
    setAssignRoleConfig({ isOpen: false, user: null });
  }

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
    handleAssignRoleAction: openAssignRoleDialog,
    handleRemoveMemberAction: openRemoveMemberDialog,
  });

  async function invalidateMembersQuery() {
    await queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: AdminServiceQueries.searchOrganizationUsers,
        transport,
        input: {},
        cardinality: "infinite",
      }),
    });
  }

  async function updateMember() {
    setAssignRoleConfig({ isOpen: false, user: null });
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
      {assignRoleConfig.isOpen && assignRoleConfig.user ? (
        <AssignRole
          roles={roles}
          user={assignRoleConfig.user}
          organizationId={organizationId}
          onRoleUpdate={updateMember}
          onClose={closeAssignRoleDialog}
        />
      ) : null}

      {removeMemberConfig.isOpen && removeMemberConfig.user ? (
        <RemoveMember
          organizationId={organizationId}
          user={removeMemberConfig.user}
          onRemove={removeMember}
          onClose={closeRemoveMemberDialog}
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
            <DataTable.Toolbar />
            <DataTable.Content
              emptyState={isError ? <ErrorState /> : <NoMembers />}
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
