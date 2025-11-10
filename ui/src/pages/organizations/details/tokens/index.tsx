import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import styles from "./tokens.module.css";
import { CoinIcon } from "@raystack/apsara/icons";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useContext, useEffect, useMemo, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import PageTitle from "~/components/page-title";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import {
  getConnectNextPageParam,
  DEFAULT_PAGE_SIZE,
} from "~/utils/connect-pagination";
import { getColumns } from "./columns";
import { useDebounceValue } from "usehooks-ts";

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
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
      heading="No tokens present"
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<CoinIcon />}
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
      heading="Error Loading Tokens"
      subHeading="Something went wrong while loading organization tokens. Please try refreshing the page."
      icon={<ExclamationTriangleIcon />}
    />
  );
};

export function OrganizationTokensPage() {
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const title = `Tokens | ${organization?.title} | Organizations`;

  const computedQuery = useMemo(() => {
    const tempQuery = transformDataTableQueryToRQLRequest(tableQuery, TRANSFORM_OPTIONS);
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
    AdminServiceQueries.searchOrganizationTokens,
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

  const columns = useMemo(() => getColumns(), []);

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
            emptyState={isError ? <ErrorState /> : <NoTokens />}
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
