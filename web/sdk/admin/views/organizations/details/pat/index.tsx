import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import { LockClosedIcon, ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useContext, useEffect, useMemo, useState } from "react";
import { useInfiniteQuery, useQuery } from "@connectrpc/connect-query";
import {
  FrontierServiceQueries,
  type Project,
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
      heading="No personal access tokens"
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<LockClosedIcon />}
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
      heading="Error Loading Personal Access Tokens"
      subHeading="Something went wrong while loading personal access tokens. Please try refreshing the page."
      icon={<ExclamationTriangleIcon />}
    />
  );
};

export function OrganizationPatView() {
  const t = useTerminology();
  const { organization, search, orgMembersMap } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

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
    FrontierServiceQueries.searchCurrentUserPATs,
    { orgId: organizationId, query },
    {
      enabled: !!organizationId,
      pageParamKey: "query",
      getNextPageParam: (lastPage) =>
        getConnectNextPageParam(lastPage, { query }, "pats"),
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

  const data = infiniteData?.pages?.flatMap((page) => page?.pats ?? []) ?? [];
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

  const columns = useMemo(
    () => getColumns({ orgMembersMap, projectsMap }),
    [orgMembersMap, projectsMap],
  );

  return (
    <Flex justify="center">
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
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={isError ? <ErrorState /> : <NoPats />}
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
