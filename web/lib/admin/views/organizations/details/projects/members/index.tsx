import { DataTable, Dialog, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery } from "@raystack/apsara";
import { useCallback, useMemo, useState } from "react";
import Skeleton from "react-loading-skeleton";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
  GetProjectRequestSchema,
  ListRolesRequestSchema,
  type SearchProjectUsersResponse_ProjectUser,
  type RQLRequest,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useQuery, useInfiniteQuery } from "@connectrpc/connect-query";
import { useDebouncedState } from "@raystack/apsara/hooks";
import styles from "./members.module.css";
import { UsersIcon } from "../../../../../assets/icons/UsersIcon";
import { getColumns } from "./columns";
import { AssignRole } from "./assign-role";
import { PROJECT_NAMESPACE } from "../../types";
import { RemoveMember } from "./remove-member";
import { AddMembersDropdown } from "./add-members-dropdown";
import {
  getConnectNextPageParam,
  DEFAULT_PAGE_SIZE,
  transformDataTableQueryToRQLRequest,
} from "@raystack/frontier/admin";

const NoMembers = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Members found"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<UsersIcon />}
    />
  );
};

const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};

export const ProjectMembersDialog = ({
  projectId,
  onClose,
}: {
  projectId: string;
  onClose: () => void;
}) => {
  const [tableQuery, setTableQuery] = useDebouncedState<{
    query: DataTableQuery;
    rqlRequest: RQLRequest;
  }>(
    {
      query: INITIAL_QUERY,
      rqlRequest: transformDataTableQueryToRQLRequest(INITIAL_QUERY, {}),
    },
    200,
  );

  const [assignRoleConfig, setAssignRoleConfig] = useState<{
    isOpen: boolean;
    user: SearchProjectUsersResponse_ProjectUser | null;
  }>({ isOpen: false, user: null });

  const [removeMemberConfig, setRemoveMemberConfig] = useState<{
    isOpen: boolean;
    user: SearchProjectUsersResponse_ProjectUser | null;
  }>({ isOpen: false, user: null });

  const { data: project, isLoading: isProjectLoading, error: projectError } = useQuery(
    FrontierServiceQueries.getProject,
    create(GetProjectRequestSchema, { id: projectId }),
    {
      enabled: !!projectId,
      select: (data) => data?.project,
    }
  );

  const { data: projectRoles = [], isLoading: isProjectRolesLoading, error: rolesError } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, { scopes: [PROJECT_NAMESPACE] }),
    {
      select: (data) => data?.roles || [],
    }
  );

  // Log errors if they occur
  if (projectError) {
    console.error("Failed to fetch project:", projectError);
  }
  if (rolesError) {
    console.error("Failed to fetch project roles:", rolesError);
  }

  const {
    data: infiniteData,
    isLoading: isMembersLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    refetch,
  } = useInfiniteQuery(
    AdminServiceQueries.searchProjectUsers,
    { id: projectId, query: tableQuery.rqlRequest },
    {
      pageParamKey: "query",
      getNextPageParam: (lastPage) =>
        getConnectNextPageParam(
          lastPage,
          { query: tableQuery.rqlRequest },
          "projectUsers",
        ),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const data =
    infiniteData?.pages?.flatMap((page) => page?.projectUsers || []) || [];

  const onTableQueryChange = useCallback((query: DataTableQuery) => {
    const updatedQuery = {
      ...query,
      offset: 0,
      limit: query.limit || DEFAULT_PAGE_SIZE,
      sort: undefined, // Remove sort as it's not supported by this endpoint
    };
    const updatedRQLRequest = transformDataTableQueryToRQLRequest(
      updatedQuery,
      {},
    );
    setTableQuery({
      query: updatedQuery,
      rqlRequest: updatedRQLRequest,
    });
  }, []);

  const handleLoadMore = useCallback(async () => {
    try {
      if (!hasNextPage) return;
      await fetchNextPage();
    } catch (error) {
      console.error("Error loading more project members:", error);
    }
  }, [hasNextPage, fetchNextPage]);

  async function refetchMembers() {
    await refetch();
  }

  const openAssignRoleDialog = useCallback(
    (user: SearchProjectUsersResponse_ProjectUser) => {
      setAssignRoleConfig({ isOpen: true, user });
    },
    [],
  );

  const closeAssignRoleDialog = useCallback(() => {
    setAssignRoleConfig({ isOpen: false, user: null });
  }, []);

  const openRemoveMemberDialog = useCallback(
    (user: SearchProjectUsersResponse_ProjectUser) => {
      setRemoveMemberConfig({ isOpen: true, user });
    },
    [],
  );

  const closeRemoveMemberDialog = useCallback(() => {
    setRemoveMemberConfig({ isOpen: false, user: null });
  }, []);

  const columns = useMemo(
    () =>
      getColumns({
        roles: projectRoles,
        handleAssignRoleAction: openAssignRoleDialog,
        handleRemoveAction: openRemoveMemberDialog,
      }),
    [projectRoles, openAssignRoleDialog, openRemoveMemberDialog],
  );

  async function removeMember(user: SearchProjectUsersResponse_ProjectUser) {
    await refetch();
    setRemoveMemberConfig({ isOpen: false, user: null });
  }

  async function updateMember(user: SearchProjectUsersResponse_ProjectUser) {
    await refetch();
    setAssignRoleConfig({ isOpen: false, user: null });
  }

  const loading = isMembersLoading || isFetchingNextPage;
  const isLoading = loading || isProjectLoading || isProjectRolesLoading;

  return (
    <>
      {assignRoleConfig.isOpen && assignRoleConfig.user ? (
        <AssignRole
          roles={projectRoles}
          user={assignRoleConfig.user}
          projectId={projectId}
          onRoleUpdate={updateMember}
          onClose={closeAssignRoleDialog}
        />
      ) : null}
      {removeMemberConfig.isOpen && removeMemberConfig.user ? (
        <RemoveMember
          projectId={projectId}
          user={removeMemberConfig.user}
          onRemove={removeMember}
          onClose={closeRemoveMemberDialog}
        />
      ) : null}
      <Dialog open onOpenChange={onClose}>
        <Dialog.Content className={styles["dialog-content"]}>
          <Dialog.Header>
            {isProjectLoading ? (
              <Skeleton containerClassName={styles["flex1"]} width={"200px"} />
            ) : (
              <Dialog.Title>{project?.title ?? ""}</Dialog.Title>
            )}
            <Dialog.CloseButton data-test-id="close-button" />
          </Dialog.Header>
          <Dialog.Body className={styles["dialog-body"]}>
            <DataTable
              query={tableQuery.query}
              columns={columns}
              data={data}
              isLoading={isLoading}
              mode="server"
              defaultSort={{ name: "", order: "desc" }}
              onTableQueryChange={onTableQueryChange}
              onLoadMore={handleLoadMore}
            >
              <Flex
                direction="column"
                gap={5}
                className={styles["table-content-wrapper"]}
              >
                <Flex>
                  <DataTable.Search className={styles["table-search"]} />
                  <AddMembersDropdown
                    projectId={projectId}
                    refetchMembers={refetchMembers}
                  />
                </Flex>
                <DataTable.Content
                  emptyState={<NoMembers />}
                  classNames={{
                    table: styles["table"],
                    root: styles["table-wrapper"],
                    header: styles["table-header"],
                  }}
                />
              </Flex>
            </DataTable>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
