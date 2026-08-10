import {
  DataTable,
  EmptyState,
  Flex,
  type DataTableQuery,
  type DataTableSort,
} from "@raystack/apsara";
import { PageTitle } from "~/admin/components/PageTitle";
import styles from "./projects.module.css";
import { useContext, useEffect, useMemo, useState } from "react";
import { getColumns } from "./columns";
import type { SearchOrganizationProjectsResponse_OrganizationProject } from "@raystack/proton/frontier";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { OrganizationContext } from "../contexts/organization-context";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { ProjectsIcon } from "~/admin/assets/icons/ProjectsIcon";
import { ProjectMembersDialog } from "./members";
import {
  getConnectNextPageParam,
  DEFAULT_PAGE_SIZE
} from '~/utils/connect-pagination';
import { transformDataTableQueryToRQLRequest } from '~/utils/transform-query';
import { useDebouncedValue } from '~hooks';
import { useTerminology } from "~/admin/hooks/useTerminology";
import { useOrgMembersMap } from "~/admin/hooks/useOrgMembersMap";

const DEFAULT_SORT: DataTableSort = { name: 'createdAt', order: 'desc' };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
  // Seeded so DataTable's mount emit matches this, instead of forcing a refetch.
  sort: [DEFAULT_SORT],
};
const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    createdAt: "created_at",
    updatedAt: "updated_at",
    userIds: "user_ids",
  },
};

const NoProjects = () => {
  const t = useTerminology();
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading={`No ${t.project({ plural: true, case: "lower" })} found`}
      subHeading="We couldn't find any matches for that keyword. Try alternative terms or check for typos."
      icon={<ProjectsIcon />}
    />
  );
};

const ZeroState = () => {
  const t = useTerminology();
  return (
    <div className={styles["zero-state-container"]}>
      <EmptyState
        variant="empty2"
        icon={<ProjectsIcon />}
        heading={t.project({ case: "capital" })}
        subHeading="A project is a structured initiative undertaken to achieve a specific outcome. It operates within a defined scope, objectives, and resources, following a process of planning, execution, monitoring, and completion."
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
      heading={`Error Loading ${t.project({ plural: true, case: "capital" })}`}
      subHeading={`Something went wrong while loading ${t.organization({ case: "lower" })} ${t.project({ plural: true, case: "lower" })}. Please try refreshing the page.`}
      icon={<ExclamationTriangleIcon />}
    />
  );
};

export function OrganizationProjectsView() {
  const t = useTerminology();
  const { organization, search } = useContext(OrganizationContext);
  const {
    data: orgMembersMap = {},
    isLoading: isOrgMembersMapLoading,
  } = useOrgMembersMap(organization?.id);
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const organizationId = organization?.id || "";

  const [memberDialogConfig, setMemberDialogConfig] = useState({
    open: false,
    projectId: "",
  });

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const title = `${t.project({ plural: true, case: "capital" })} | ${organization?.title} | ${t.organization({ plural: true, case: "capital" })}`;

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
    refetch: refetchOrgProjects,
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

  /*
   * DataTable seeds its query once at mount, so it never sees the org-context
   * search. Hence picking the state here instead of via the zeroState prop.
   */
  const hasActiveQuery = Boolean(
    searchQuery?.trim() || tableQuery.filters?.length,
  );
  const showZeroState =
    !isLoading && !isError && !hasActiveQuery && data.length === 0;

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
    // Refetch the query instead of manually updating
    refetchOrgProjects();
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
    refetchOrgProjects();
    setMemberDialogConfig({
      projectId: "",
      open: false,
    });
  }

  const columns = getColumns({ orgMembersMap, handleProjectUpdate, t });

  const canAddMember = Object.keys(orgMembersMap).length > 1;

  return (
    <>
      {memberDialogConfig.open && memberDialogConfig.projectId ? (
        <ProjectMembersDialog
          projectId={memberDialogConfig.projectId}
          onClose={handleMemberDialogClose}
          canAddMember={canAddMember}
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
              emptyState={showZeroState ? <ZeroState /> : isError ? <ErrorState /> : <NoProjects />}
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
