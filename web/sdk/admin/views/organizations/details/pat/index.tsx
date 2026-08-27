import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableSort } from "@raystack/apsara";
import { LockClosedIcon, ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useInfiniteQuery, useQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
  type Project,
  type SearchOrganizationPATsResponse_OrganizationPAT,
} from "@raystack/proton/frontier";
import { OrganizationContext } from "../contexts/organization-context";
import { PageTitle } from "~/admin/components/PageTitle";
import {
  getConnectNextPageParam,
} from "~/utils/connect-pagination";
import { useTerminology } from "~/admin/hooks/useTerminology";
import { getColumns } from "./columns";
import { PatDetailsDialog } from "./components/pat-details-dialog";
import styles from "./pat.module.css";
import { useLoadMore } from "~/admin/hooks/useLoadMore";
import { useServerTableQuery } from "~/admin/hooks/useServerTableQuery";

const DEFAULT_SORT: DataTableSort = { name: "createdAt", order: "desc" };
const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    createdAt: "created_at",
    updatedAt: "updated_at",
    expiresAt: "expires_at",
    usedAt: "used_at",
    createdBy: "created_by_title",
  },
};

const NoPats = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No PAT found"
      subHeading="We couldn't find any matches for that keyword. Try alternative terms or check for typos."
      icon={<LockClosedIcon />}
    />
  );
};

const ZeroState = () => {
  return (
    <div className={styles["zero-state-container"]}>
      <EmptyState
        variant="empty2"
        icon={<LockClosedIcon />}
        heading="PAT"
        subHeading="A Personal Access Token (PAT) is a secure credential that allows external applications and scripts to interact with Aurora APIs. It enables authenticated access to resources and workflows without requiring direct user login."
      />
    </div>
  );
};

const ErrorState = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="Error Loading Personal Access Tokens"
      subHeading="Something went wrong while loading personal access tokens. Please try refreshing the page."
      icon={<ExclamationTriangleIcon />}
    />
  );
};

export function OrganizationPatView() {
  const t = useTerminology();
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const [selectedPat, setSelectedPat] =
    useState<SearchOrganizationPATsResponse_OrganizationPAT | null>(null);

  const title = `PAT | ${organization?.title} | ${t.organization({ plural: true, case: "capital" })}`;

  const {
    tableQuery,
    rqlQuery: query,
    onTableQueryChange,
  } = useServerTableQuery({
    defaultSort: DEFAULT_SORT,
    transformOptions: TRANSFORM_OPTIONS,
    search: searchQuery || "",
  });

  const {
    data: infiniteData,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
    isError,
  } = useInfiniteQuery(
    AdminServiceQueries.searchOrganizationPATs,
    { orgId: organizationId, query },
    {
      enabled: !!organizationId,
      pageParamKey: "query",
      getNextPageParam: (lastPage) =>
        getConnectNextPageParam(lastPage, { query }, "organizationPats"),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const { data: projects = [] } = useQuery(
    FrontierServiceQueries.listOrganizationProjects,
    { id: organizationId, state: "", withMemberCount: false },
    {
      enabled: !!organizationId,
      select: (data) => data?.projects || [],
    },
  );

  const projectsMap = useMemo(
    () =>
      projects.reduce(
        (acc, project) => {
          if (project.id) acc[project.id] = project;
          return acc;
        },
        {} as Record<string, Project>,
      ),
    [projects],
  );

  const data = useMemo(
    () =>
      infiniteData?.pages?.flatMap((page) => page?.organizationPats ?? []) ?? [],
    [infiniteData],
  );
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

  const fetchMore = useLoadMore({
    hasNextPage,
    isFetchingNextPage,
    isError,
    fetchNextPage,
    label: "personal access tokens",
  });

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  const columns = useMemo(
    () => getColumns({ projectsMap }),
    [projectsMap],
  );

  const onRowClick = useCallback(
    (row: SearchOrganizationPATsResponse_OrganizationPAT) => {
      setSelectedPat(row);
    },
    [],
  );

  const onDialogClose = useCallback(() => {
    setSelectedPat(null);
  }, []);

  return (
    <Flex justify="center">
      <PageTitle title={title} />
      <PatDetailsDialog
        pat={selectedPat}
        projectsMap={projectsMap}
        onClose={onDialogClose}
      />
      <DataTable
        columns={columns}
        data={data}
        isLoading={loading}
        defaultSort={DEFAULT_SORT}
        mode="server"
        onTableQueryChange={onTableQueryChange}
        onLoadMore={fetchMore}
        onRowClick={onRowClick}
        query={tableQuery}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={showZeroState ? <ZeroState /> : isError ? <ErrorState /> : <NoPats />}
            classNames={{
              table: styles["table"],
              root: styles["table-wrapper"],
              header: styles["table-header"],
            }}
          />
        </Flex>
      </DataTable>
    </Flex>
  );
}
