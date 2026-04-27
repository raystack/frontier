import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import { LockClosedIcon, ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useInfiniteQuery, useQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
  type Project,
  type SearchOrganizationPATsResponse_OrganizationPAT,
} from "@raystack/proton/frontier";
import { useDebounceValue } from "usehooks-ts";
import { OrganizationContext } from "../contexts/organization-context";
import { PageTitle } from "../../../../components/PageTitle";
import {
  DEFAULT_PAGE_SIZE,
  getConnectNextPageParam,
} from "~/utils/connect-pagination";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import { useTerminology } from "../../../../hooks/useTerminology";
import { getColumns } from "./columns";
import { PatDetailsDialog } from "./components/pat-details-dialog";
import styles from "./pat.module.css";

const DEFAULT_SORT: DataTableSort = { name: "createdAt", order: "desc" };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};
const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    createdAt: "created_at",
    updatedAt: "updated_at",
    expiresAt: "expires_at",
    usedAt: "used_at",
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
        subHeading="Personal access tokens (PATs) provide programmatic access to organization resources via the API on behalf of a user."
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

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);
  const [selectedPat, setSelectedPat] =
    useState<SearchOrganizationPATsResponse_OrganizationPAT | null>(null);

  const title = `PAT | ${organization?.title} | ${t.organization({ plural: true, case: "capital" })}`;

  const computedQuery = useMemo(() => {
    const tempQuery = transformDataTableQueryToRQLRequest(
      tableQuery,
      TRANSFORM_OPTIONS,
    );
    return {
      ...tempQuery,
      search: searchQuery || "",
    };
  }, [tableQuery, searchQuery]);

  const [query] = useDebounceValue(computedQuery, 200);

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

  const data =
    infiniteData?.pages?.flatMap((page) => page?.organizationPats ?? []) ?? [];
  const loading = (isLoading || isFetchingNextPage) && !isError;

  const hasActiveQuery = Boolean(query.search || query.filters?.length);
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
