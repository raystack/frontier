import {
  DataTable,
  type DataTableQuery,
  type DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara";
import { useDebouncedState } from "@raystack/apsara/hooks";
import { useCallback, useMemo, useState } from "react";
import Navbar from "./navbar";
import styles from "./audit-logs.module.css";
import { getColumns } from "./columns";
import { PageTitle } from "../../components/PageTitle";
import { CpuChipIcon } from "../../assets/icons/CpuChipIcon";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  AuditRecord,
  RQLRequest,
} from "@raystack/proton/frontier";
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage,
  DEFAULT_PAGE_SIZE,
} from "../../utils/connect-pagination";
import { transformDataTableQueryToRQLRequest } from "../../utils/transform-query";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import SidePanelDetails from "./sidepanel-details";
import { useQueryClient } from "@tanstack/react-query";
import { AUDIT_LOG_QUERY_KEY } from "./util";

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
const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    occurredAt: "occurred_at",
    orgId: "org_id",
    orgName: "org_name",
    actor: "actor_name",
    resource: "resource_name",
    resourceType: "resource_type",
    actorType: "actor_type",
  },
};

export type AuditLogsViewProps = {
  appName?: string;
  onExportCsv?: (query: RQLRequest) => Promise<void>;
};

export default function AuditLogsView({ appName, onExportCsv }: AuditLogsViewProps = {}) {
  const queryClient = useQueryClient();
  const [tableQuery, setTableQuery] = useDebouncedState<{
    query: DataTableQuery;
    rqlRequest: RQLRequest;
  }>(
    {
      query: INITIAL_QUERY,
      rqlRequest: transformDataTableQueryToRQLRequest(
        INITIAL_QUERY,
        TRANSFORM_OPTIONS,
      ),
    },
    200,
  );
  const [sidePanelOpen, setSidePanelOpen] = useState(false);
  const [selectedAuditLog, setSelectedAuditLog] = useState<AuditRecord | null>(
    null,
  );

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
    { query: tableQuery.rqlRequest },
    {
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(
          lastPage,
          { query: tableQuery.rqlRequest },
          "auditRecords",
        ),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const data =
    infiniteData?.pages?.flatMap(page => page?.auditRecords || []) || [];

  const onTableQueryChange = useCallback(
    (query: DataTableQuery) => {
      const updatedQuery = {
        ...query,
        offset: 0,
        limit: query.limit || DEFAULT_PAGE_SIZE,
      };
      const updatedRQLRequest = transformDataTableQueryToRQLRequest(
        updatedQuery,
        TRANSFORM_OPTIONS,
      );
      queryClient.setQueryData(AUDIT_LOG_QUERY_KEY, updatedRQLRequest);
      setTableQuery({
        query: updatedQuery,
        rqlRequest: updatedRQLRequest,
      });
    },
    [queryClient],
  );

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
        <PageTitle title="Audit Logs" appName={appName} />
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
      <PageTitle title="Audit Logs" appName={appName} />
      <DataTable
        query={tableQuery.query}
        columns={columns}
        data={data}
        isLoading={loading}
        defaultSort={DEFAULT_SORT}
        onTableQueryChange={onTableQueryChange}
        mode="server"
        onLoadMore={handleLoadMore}
        onRowClick={onRowClick}>
        <Flex direction="column" style={{ width: "100%" }}>
          <Navbar searchQuery={tableQuery.query.search} onExportCsv={onExportCsv} />
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
}
