import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import styles from "./apis.module.css";
import {
  CodeIcon,
  ExclamationTriangleIcon,
} from "@radix-ui/react-icons";
import { useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { PageTitle } from "~/admin/components/PageTitle";
import { getColumns } from "./columns";
import { ServiceUserDetailsDialog } from "./details-dialog";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  type SearchOrganizationServiceUsersResponse_OrganizationServiceUser,
} from "@raystack/proton/frontier";
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage,
  DEFAULT_PAGE_SIZE,
} from "~/utils/connect-pagination";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import { useDebouncedValue } from "~hooks";
import { useTerminology } from "~/admin/hooks/useTerminology";

const NoCredentials = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No service account found"
      subHeading="We couldn't find any matches for that keyword. Try alternative terms or check for typos."
      icon={<CodeIcon />}
    />
  );
};

const ZeroState = () => {
  return (
    <div className={styles["zero-state-container"]}>
      <EmptyState
        variant="empty2"
        icon={<CodeIcon />}
        heading="API"
        subHeading="An API is a set of protocols that enables Aurora to interact with other applications. It defines request and response structures, allowing seamless data exchange and integration."
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
      heading="Error Loading API Credentials"
      subHeading="Something went wrong while loading API credentials. Please try refreshing the page."
      icon={<ExclamationTriangleIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: 'createdAt', order: 'desc' };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
  // Must match DataTable's mount emit, or it refetches.
  sort: [DEFAULT_SORT],
};
const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    createdAt: "created_at",
  },
};

export function OrganizationApisView() {
  const t = useTerminology();
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const computedQuery = useMemo(() => {
    const tempQuery = transformDataTableQueryToRQLRequest(tableQuery, TRANSFORM_OPTIONS);
    return {
      ...tempQuery,
      search: searchQuery || "",
    };
  }, [tableQuery, searchQuery]);

  const query = useDebouncedValue(computedQuery, 200);


  const [selectedServiceUser, setSelectedServiceUser] =
    useState<SearchOrganizationServiceUsersResponse_OrganizationServiceUser | null>(
      null,
    );

  const title = `API | ${organization?.title} | ${t.organization({ plural: true, case: "capital" })}`;

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  const {
    data: infiniteData,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    error,
    isError,
    hasNextPage,
  } = useInfiniteQuery(
    AdminServiceQueries.searchOrganizationServiceUsers,
    { id: organizationId, query: query },
    {
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(
          lastPage,
          { query: query },
          "organizationServiceUsers",
        ),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
      enabled: !!organizationId,
    },
  );

  const data =
    infiniteData?.pages?.flatMap(page => page?.organizationServiceUsers || []) || [];

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery(newQuery);
  };

  // isFetchingNextPage lags a render; the ref doesn't.
  const isLoadingMoreRef = useRef(false);
  const handleLoadMore = async () => {
    if (!hasNextPage || isFetchingNextPage || isLoadingMoreRef.current) return;
    isLoadingMoreRef.current = true;
    try {
      await fetchNextPage();
    } catch (error) {
      console.error("Error loading more service users:", error);
    } finally {
      isLoadingMoreRef.current = false;
    }
  };

  const loading = isLoading || isFetchingNextPage;

  /*
   * DataTable seeds its query once at mount, so it never sees the org-context
   * search. Hence picking the state here instead of via the zeroState prop.
   */
  const hasActiveQuery = Boolean(
    searchQuery?.trim() || tableQuery.filters?.length,
  );
  const showZeroState =
    !isLoading && !isError && !hasActiveQuery && data.length === 0;

  const onDialogClose = useCallback(() => {
    setSelectedServiceUser(null);
  }, []);

  const onRowClick = useCallback(
    (row: SearchOrganizationServiceUsersResponse_OrganizationServiceUser) => {
      setSelectedServiceUser(row);
    },
    [],
  );

  if (isError) {
    console.error("ConnectRPC Error:", error);
  }

  const columns = useMemo(
    () =>
      getColumns({
        groupCountMap: infiniteData
          ? getGroupCountMapFromFirstPage(infiniteData)
          : {},
        t,
      }),
    [infiniteData, t],
  );

  return (
    <Flex justify="center" className={styles["container"]}>
      <ServiceUserDetailsDialog
        serviceUser={selectedServiceUser}
        onClose={onDialogClose}
      />
      <PageTitle title={title} />
      <DataTable
        query={tableQuery}
        columns={columns}
        data={data}
        isLoading={loading}
        defaultSort={DEFAULT_SORT}
        onTableQueryChange={onTableQueryChange}
        mode="server"
        onLoadMore={handleLoadMore}
        onRowClick={onRowClick}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={showZeroState ? <ZeroState /> : isError ? <ErrorState /> : <NoCredentials />}
            classNames={{
              root: styles["table-wrapper"],
              table: styles["table"],
              header: styles["table-header"],
            }}
          />
        </Flex>
      </DataTable>
    </Flex>
  );
}
