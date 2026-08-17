import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import styles from "./tokens.module.css";
import { CoinIcon } from "@raystack/apsara/icons";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useContext, useEffect, useMemo, useRef, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { PageTitle } from "~/admin/components/PageTitle";
import { FrontierServiceQueries } from "@raystack/proton/frontier";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { getConnectNextPageParam, DEFAULT_PAGE_SIZE } from "~/utils/connect-pagination";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import { getColumns } from "./columns";
import { useDebouncedValue } from "~hooks";
import { useTerminology } from "~/admin/hooks/useTerminology";

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
    updatedAt: "updated_at",
    expiresAt: "expires_at",
    transactionId: "transaction_id",
    userId: "user_id",
    userTitle: "user_title",
    userAvatar: "user_avatar",
  },
};

const NoTokens = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No token present"
      subHeading="We couldn't find any matches for that keyword. Try alternative terms or check for typos."
      icon={<CoinIcon />}
    />
  );
};

const ZeroState = () => {
  return (
    <div className={styles["zero-state-container"]}>
      <EmptyState
        variant="empty2"
        icon={<CoinIcon />}
        heading="Tokens"
        subHeading="Tokens serve as a flexible currency, allowing organizations to access services like satellite imagery orders and advanced analytics. They provide a scalable way to manage resources and adapt to evolving needs."
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
      heading="Error Loading Tokens"
      subHeading={`Something went wrong while loading ${t.organization({ case: "lower" })} tokens. Please try refreshing the page.`}
      icon={<ExclamationTriangleIcon />}
    />
  );
};

export function OrganizationTokensView() {
  const t = useTerminology();
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const title = `Tokens | ${organization?.title} | ${t.organization({ plural: true, case: "capital" })}`;

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
    FrontierServiceQueries.searchOrganizationTokens,
    { id: organizationId, query: query },
    {
      enabled: !!organizationId,
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(
          lastPage,
          { query: query },
          "organizationTokens",
        ),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const data =
    infiniteData?.pages?.flatMap(page => page.organizationTokens) || [];
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
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  const columns = useMemo(() => getColumns({ t }), [t]);

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
        query={tableQuery}>
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={showZeroState ? <ZeroState /> : isError ? <ErrorState /> : <NoTokens />}
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
