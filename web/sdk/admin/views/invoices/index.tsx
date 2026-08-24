import {
  DataTable,
  type DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara";
import { PageTitle } from "../../components/PageTitle";
import { InvoicesNavabar } from "./navbar";
import styles from "./invoices.module.css";
import { InvoicesIcon } from "../../assets/icons/InvoicesIcon";
import { getColumns } from "./columns";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import {
  getConnectNextPageParam,
} from "~/utils/connect-pagination";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useTerminology } from "../../hooks/useTerminology";
import { useLoadMore } from "~/admin/hooks/useLoadMore";
import { useServerTableQuery } from "~/admin/hooks/useServerTableQuery";

const NoInvoices = () => {
  const t = useTerminology();
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No invoices found"
      subHeading={`Start billing to ${t.organization({ plural: true, case: "lower" })} to populate the table`}
      icon={<InvoicesIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "createdAt", order: "desc" };

export type InvoicesViewProps = {
  /** App name displayed in the page title. */
  appName?: string;
};

export default function InvoicesView({ appName }: InvoicesViewProps = {}) {
  const t = useTerminology();
  const {
    tableQuery,
    rqlQuery: query,
    onTableQueryChange,
  } = useServerTableQuery({
    defaultSort: DEFAULT_SORT,
    transformOptions: {
      fieldNameMapping: {
        createdAt: "created_at",
      },
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

  const handleLoadMore = useLoadMore({
    hasNextPage,
    isFetchingNextPage,
    isError,
    fetchNextPage,
    label: "invoices",
  });

  const columns = getColumns({ t });

  const loading = isLoading || isFetchingNextPage;

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <>
        <PageTitle title="Invoices" appName={appName} />
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
      <PageTitle title="Invoices" appName={appName} />
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
          <DataTable.VirtualizedContent
            classNames={{
              root: styles["table-wrapper"],
              table: tableClassName,
              header: styles["table-header"],
            }}
            emptyState={<NoInvoices />}
            rowHeight={48}
            groupHeaderHeight={48}
          />
        </Flex>
      </DataTable>
    </>
  );
}
