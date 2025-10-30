import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import PageTitle from "~/components/page-title";
import styles from "./projects.module.css";
import { useContext, useEffect, useState } from "react";
import { getColumns } from "./columns";
import type { SearchOrganizationProjectsResponse_OrganizationProject } from "@raystack/proton/frontier";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import {
  useInfiniteQuery,
  createConnectQueryKey,
  useTransport,
} from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { OrganizationContext } from "../contexts/organization-context";
import { FileIcon, ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { ProjectMembersDialog } from "./members";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import {
  getConnectNextPageParam,
  DEFAULT_PAGE_SIZE,
} from "~/utils/connect-pagination";
import { useDebouncedState } from "~/hooks/useDebouncedState";

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};

const NoProjects = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Projects found"
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<FileIcon />}
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
      heading="Error Loading Projects"
      subHeading="Something went wrong while loading organization projects. Please try refreshing the page."
      icon={<ExclamationTriangleIcon />}
    />
  );
};

export function OrganizationProjectssPage() {
  const { organization, search, orgMembersMap, isOrgMembersMapLoading } =
    useContext(OrganizationContext);
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;
  const queryClient = useQueryClient();
  const transport = useTransport();

  const organizationId = organization?.id || "";

  const [memberDialogConfig, setMemberDialogConfig] = useState({
    open: false,
    projectId: "",
  });

  const [tableQuery, setTableQuery] = useDebouncedState<DataTableQuery>(
    INITIAL_QUERY,
    200,
  );

  const title = `Projects | ${organization?.title} | Organizations`;

  // Transform the DataTableQuery to RQLRequest format
  const query = transformDataTableQueryToRQLRequest(tableQuery, {
    fieldNameMapping: {
      createdAt: "created_at",
      updatedAt: "updated_at",
      userIds: "user_ids",
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
    AdminServiceQueries.searchOrganizationProjects,
    { id: organizationId, query: query },
    {
      enabled: !!organizationId,
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(lastPage, { query: query }, "orgProjects"),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const data = infiniteData?.pages?.flatMap(page => page.orgProjects) || [];
  const loading =
    (isLoading || isFetchingNextPage || isOrgMembersMapLoading) && !isError;

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery(newQuery);
  };

  const fetchMore = async () => {
    if (hasNextPage && !isFetchingNextPage && !isError) {
      await fetchNextPage();
    }
  };

  function handleProjectUpdate(
    project: SearchOrganizationProjectsResponse_OrganizationProject,
  ) {
    // Invalidate and refetch the query instead of manually updating
    queryClient.invalidateQueries({
      queryKey: createConnectQueryKey({
        schema: AdminServiceQueries.searchOrganizationProjects,
        transport,
        input: {},
        cardinality: "infinite",
      }),
    });
  }

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  function handleMemberDialogOpen(
    project: SearchOrganizationProjectsResponse_OrganizationProject,
  ) {
    setMemberDialogConfig({
      projectId: project.id || "",
      open: true,
    });
  }

  function handleMemberDialogClose() {
    setMemberDialogConfig({
      projectId: "",
      open: false,
    });
  }

  const columns = getColumns({ orgMembersMap, handleProjectUpdate });

  return (
    <>
      {memberDialogConfig.open && memberDialogConfig.projectId ? (
        <ProjectMembersDialog
          projectId={memberDialogConfig.projectId}
          onClose={handleMemberDialogClose}
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
          query={tableQuery}
          onRowClick={handleMemberDialogOpen}>
          <Flex direction="column" style={{ width: "100%" }}>
            <DataTable.Toolbar />
            <DataTable.Content
              emptyState={isError ? <ErrorState /> : <NoProjects />}
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
