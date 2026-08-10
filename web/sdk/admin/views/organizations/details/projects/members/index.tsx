import { AlertDialog, DataTable, Dialog, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery } from "@raystack/apsara";
import { useCallback, useMemo, useState } from "react";
import Skeleton from "react-loading-skeleton";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
  GetProjectRequestSchema,
  ListRolesRequestSchema,
  type SearchProjectUsersResponse_ProjectUser,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useQuery, useInfiniteQuery } from "@connectrpc/connect-query";
import styles from "./members.module.css";
import { UsersIcon } from "../../../../../assets/icons/UsersIcon";
import { getColumns } from "./columns";
import { UpdateRole, type UpdateRolePayload } from "./update-role";
import { PROJECT_NAMESPACE } from "../../types";
import { RemoveMember } from "./remove-member";
import { AddMembersDropdown } from "./add-members-dropdown";
import { getConnectNextPageParam } from "~/utils/connect-pagination";
import { useServerTableQuery } from "~/admin/hooks/useServerTableQuery";

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

/* This endpoint ignores sort, so keep it out of the request. */
const stripSort = (query: DataTableQuery): DataTableQuery => ({
  ...query,
  sort: [],
});

const updateRoleDialogHandle = AlertDialog.createHandle<UpdateRolePayload>();

export const ProjectMembersDialog = ({
  projectId,
  onClose,
  canAddMember,
}: {
  projectId: string;
  onClose: () => void;
  canAddMember: boolean;
}) => {
  const { tableQuery, rqlQuery, onTableQueryChange } = useServerTableQuery({
    mapQuery: stripSort,
  });

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
    { id: projectId, query: rqlQuery },
    {
      pageParamKey: "query",
      getNextPageParam: (lastPage) =>
        getConnectNextPageParam(
          lastPage,
          { query: rqlQuery },
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

  const handleLoadMore = useCallback(async () => {
    if (!hasNextPage || isFetchingNextPage) return;
    try {
      await fetchNextPage();
    } catch (error) {
      console.error("Error loading more project members:", error);
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  async function refetchMembers() {
    await refetch();
  }

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
        updateRoleHandle: updateRoleDialogHandle,
        handleRemoveAction: openRemoveMemberDialog,
      }),
    [projectRoles, openRemoveMemberDialog],
  );

  async function removeMember(user: SearchProjectUsersResponse_ProjectUser) {
    await refetch();
    setRemoveMemberConfig({ isOpen: false, user: null });
  }

  async function updateMember() {
    await refetch();
  }

  const loading = isMembersLoading || isFetchingNextPage;
  const isLoading = loading || isProjectLoading || isProjectRolesLoading;

  return (
    <>
      <UpdateRole
        handle={updateRoleDialogHandle}
        projectId={projectId}
        onRoleUpdate={updateMember}
      />
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
              query={tableQuery}
              columns={columns}
              data={data}
              isLoading={isLoading}
              mode="server"
              onTableQueryChange={onTableQueryChange}
              onLoadMore={handleLoadMore}
            >
              <Flex
                direction="column"
                gap={5}
                className={styles["table-content-wrapper"]}
              >
                <Flex gap={4} align="center">
                  <DataTable.Search className={styles["table-search"]} />
                  <AddMembersDropdown
                    projectId={projectId}
                    refetchMembers={refetchMembers}
                    disabled={!canAddMember}
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
