import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import { useCallback, useMemo, useState } from "react";
import Navbar from "./navbar";
import styles from "./list.module.css";
import { getColumns } from "./columns";
import PageTitle from "~/components/page-title";
import CpuChipIcon from "~/assets/icons/cpu-chip.svg?react";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries, AuditRecord } from "@raystack/proton/frontier";
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage,
  DEFAULT_PAGE_SIZE,
} from "~/utils/connect-pagination";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import SidePanelDetails from "./sidepanel-details";

const NoAuditLogs = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No activity logged yet"
      subHeading="Once users start making changes, you'll see a detailed history of events here."
      icon={<CpuChipIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "occurredAt", order: "desc" };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};

export const AuditLogsList = () => {
  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);
  const [sidePanelOpen, setSidePanelOpen] = useState(false);
  const [selectedAuditLog, setSelectedAuditLog] = useState<AuditRecord | null>(
    null,
  );

  const query = transformDataTableQueryToRQLRequest(tableQuery, {
    fieldNameMapping: {
      occurredAt: "occurred_at",
      orgId: "org_id",
      actor: "actor.name",
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
    AdminServiceQueries.listAuditRecords,
    { query: query },
    {
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(lastPage, { query: query }, "auditRecords"),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const data =
    infiniteData?.pages?.flatMap(page => page?.auditRecords || []) || [];

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
      console.error("Error loading more audit logs:", error);
    }
  };

  const columns = useMemo(
    () =>
      getColumns({
        groupCountMap: infiniteData
          ? getGroupCountMapFromFirstPage(infiniteData)
          : {},
      }),
    [infiniteData],
  );

  const loading = isLoading || isFetchingNextPage;

  const onRowClick = useCallback((row: AuditRecord) => {
    setSelectedAuditLog(_selectedLog => {
      if (_selectedLog?.id === row.id) setSidePanelOpen(_value => !_value);
      else setSidePanelOpen(true);
      return row;
    });
  }, []);

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <>
        <PageTitle title="Audit Logs" />
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="Error Loading Audit Logs"
          subHeading={
            error?.message ||
            "Something went wrong while loading audit logs. Please try again."
          }
        />
      </>
    );
  }

  const tableClassName =
    data.length || loading ? styles["table"] : styles["table-empty"];

  return (
    <>
      <PageTitle title="Audit Logs" />
      <DataTable
        query={tableQuery}
        columns={columns}
        data={data}
        isLoading={loading}
        defaultSort={DEFAULT_SORT}
        onTableQueryChange={onTableQueryChange}
        mode="server"
        onLoadMore={handleLoadMore}
        onRowClick={onRowClick}>
        <Flex direction="column" style={{ width: "100%" }}>
          <Navbar searchQuery={tableQuery.search} />
          <DataTable.Toolbar />
          <Flex className={styles["table-content-container"]}>
            <DataTable.Content
              classNames={{
                root: styles["table-wrapper"],
                table: tableClassName,
                header: styles["table-header"],
              }}
              emptyState={<NoAuditLogs />}
            />
            {sidePanelOpen && (
              <SidePanelDetails
                {...selectedAuditLog}
                onClose={() => setSidePanelOpen(false)}
              />
            )}
          </Flex>
        </Flex>
      </DataTable>
    </>
  );
};
