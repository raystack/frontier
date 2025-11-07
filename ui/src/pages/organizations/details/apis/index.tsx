import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import { useDebouncedState } from "@raystack/apsara/hooks";
import styles from "./apis.module.css";
import { InfoCircledIcon, ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useCallback, useContext, useEffect, useMemo, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import PageTitle from "~/components/page-title";
import { getColumns } from "./columns";
import { ServiceUserDetailsDialog } from "./details-dialog";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  type SearchOrganizationServiceUsersResponse_OrganizationServiceUser,
} from "@raystack/proton/frontier";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage,
  DEFAULT_PAGE_SIZE,
} from "~/utils/connect-pagination";
import { RQLRequest } from "@raystack/frontier";

const NoCredentials = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No API Credentials found"
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      // TODO: update icon with raystack icon
      icon={<InfoCircledIcon />}
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
      heading="Error Loading API Credentials"
      subHeading="Something went wrong while loading API credentials. Please try refreshing the page."
      icon={<ExclamationTriangleIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};
const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    createdAt: "created_at",
  },
};

export function OrganizationApisPage() {
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";
  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);
  const [query, setQuery] = useDebouncedState<RQLRequest | undefined>(undefined, 200);

  useEffect(() => {
    const tempQuery = transformDataTableQueryToRQLRequest(tableQuery, TRANSFORM_OPTIONS);
    setQuery({
      ...tempQuery,
      search: searchQuery || "",
    });
  }, [tableQuery, searchQuery, setQuery]);

  const [selectedServiceUser, setSelectedServiceUser] =
    useState<SearchOrganizationServiceUsersResponse_OrganizationServiceUser | null>(
      null,
    );

  const title = `API | ${organization?.title} | Organizations`;

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

  const handleLoadMore = async () => {
    try {
      if (!hasNextPage) return;
      await fetchNextPage();
    } catch (error) {
      console.error("Error loading more service users:", error);
    }
  };

  const loading = isLoading || isFetchingNextPage;

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
      }),
    [infiniteData],
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
            emptyState={isError ? <ErrorState /> : <NoCredentials />}
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
