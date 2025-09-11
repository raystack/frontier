import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import styles from "./invoices.module.css";
import { FileTextIcon, ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useContext, useEffect, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import PageTitle from "~/components/page-title";
import { getColumns } from "./columns";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import {
  getConnectNextPageParam,
  DEFAULT_PAGE_SIZE,
  getGroupCountMapFromFirstPage,
} from "~/utils/connect-pagination";

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};

const NoInvoices = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Invoice found"
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      // TODO: update icon with raystack icon
      icon={<FileTextIcon />}
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
      heading="Error Loading Invoices"
      subHeading="Something went wrong while loading organization invoices. Please try refreshing the page."
      icon={<ExclamationTriangleIcon />}
    />
  );
};

export function OrganizationInvoicesPage() {
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";

  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const title = `Invoices | ${organization?.title} | Organizations`;

  // Transform the DataTableQuery to RQLRequest format
  const query = transformDataTableQueryToRQLRequest(tableQuery, {
    fieldNameMapping: {
      createdAt: "created_at",
      updatedAt: "updated_at",
      dueDate: "due_date",
      paidAt: "paid_at",
      effectiveAt: "effective_at",
      periodStart: "period_start",
      periodEnd: "period_end",
      invoiceLink: "invoice_link",
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
    AdminServiceQueries.searchOrganizationInvoices,
    { id: organizationId, query: query },
    {
      enabled: !!organizationId,
      pageParamKey: "query",
      getNextPageParam: (lastPage) =>
        getConnectNextPageParam(
          lastPage,
          { query: query },
          "organizationInvoices",
        ),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const data =
    infiniteData?.pages?.flatMap((page) => page.organizationInvoices) || [];
  const loading = (isLoading || isFetchingNextPage) && !isError;
  const groupCountMap = infiniteData
    ? getGroupCountMapFromFirstPage(infiniteData)
    : {};

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

  const columns = getColumns({ groupCountMap });
  return (
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
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={isError ? <ErrorState /> : <NoInvoices />}
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
