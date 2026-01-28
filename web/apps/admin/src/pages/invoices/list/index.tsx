import {
  DataTable,
  type DataTableQuery,
  type DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara";
import { useMemo, useState } from "react";
import PageTitle from "~/components/page-title";
import { InvoicesNavabar } from "./navbar";
import styles from "./list.module.css";
import InvoicesIcon from "~/assets/icons/invoices.svg?react";
import { getColumns } from "./columns";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import {
  getConnectNextPageParam,
  DEFAULT_PAGE_SIZE,
} from "~/utils/connect-pagination";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";

const NoInvoices = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No invoices found"
      subHeading="Start billing to organizations to populate the table"
      icon={<InvoicesIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "createdAt", order: "desc" };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};

export const InvoicesList = () => {
  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const query = transformDataTableQueryToRQLRequest(tableQuery, {
    fieldNameMapping: {
      createdAt: "created_at",
    },
  });

  const {
    data: infiniteData,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    error,
    isError,
    hasNextPage,
  } = useInfiniteQuery(
    AdminServiceQueries.searchInvoices,
    { query },
    {
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(lastPage, { query: query }, "invoices"),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const data = infiniteData?.pages?.flatMap(page => page?.invoices || []) || [];

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery({
      ...newQuery,
      offset: 0,
      limit: newQuery.limit || DEFAULT_PAGE_SIZE,
    });
  };

  const handleLoadMore = async () => {
    try {
      if (!hasNextPage) return;
      await fetchNextPage();
    } catch (error) {
      console.error("Error loading more invoices:", error);
    }
  };

  const columns = getColumns();

  const loading = isLoading || isFetchingNextPage;

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <>
        <PageTitle title="Invoices" />
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="Error Loading Invoices"
          subHeading={
            error?.message ||
            "Something went wrong while loading invoices. Please try again."
          }
        />
      </>
    );
  }

  const tableClassName =
    data.length || loading ? styles["table"] : styles["table-empty"];

  return (
    <>
      <PageTitle title="Invoices" />
      <DataTable
        query={tableQuery}
        columns={columns}
        data={data}
        isLoading={loading}
        defaultSort={DEFAULT_SORT}
        onTableQueryChange={onTableQueryChange}
        mode="server"
        onLoadMore={handleLoadMore}>
        <Flex direction="column" style={{ width: "100%" }}>
          <InvoicesNavabar searchQuery={tableQuery.search || ""} />
          <DataTable.Toolbar />
          <DataTable.Content
            classNames={{
              root: styles["table-wrapper"],
              table: tableClassName,
              header: styles["table-header"],
            }}
            emptyState={<NoInvoices />}
          />
        </Flex>
      </DataTable>
    </>
  );
};
