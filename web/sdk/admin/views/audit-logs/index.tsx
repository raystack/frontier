import {
  DataTable,
  type DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara";
import { useEffect, useCallback, useMemo, useState } from "react";
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
} from "~/utils/connect-pagination";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import SidePanelDetails from "./sidepanel-details";
import { useQueryClient } from "@tanstack/react-query";
import { AUDIT_LOG_QUERY_KEY } from "./util";
import { useTerminology } from "../../hooks/useTerminology";
import { useLoadMore } from "~/admin/hooks/useLoadMore";
import { useServerTableQuery } from "~/admin/hooks/useServerTableQuery";

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
  /** App name displayed in the page title. */
  appName?: string;
  /** Callback to export audit logs as CSV with the current query filters applied. */
  onExportCsv?: (query: RQLRequest) => Promise<void>;
  /** Navigate to a link in an audit entry (e.g. org/user page). `state` carries the org id. */
  onNavigate?: (path: string, state?: { orgId?: string }) => void;
};

export default function AuditLogsView({ appName, onExportCsv, onNavigate }: AuditLogsViewProps = {}) {
  const t = useTerminology();
  const queryClient = useQueryClient();
  const {
    tableQuery,
    rqlQuery,
    onTableQueryChange,
  } = useServerTableQuery({
    defaultSort: DEFAULT_SORT,
    transformOptions: TRANSFORM_OPTIONS,
  });

  /* The navbar's CSV export reads the live request off this key. */
  useEffect(() => {
    queryClient.setQueryData(AUDIT_LOG_QUERY_KEY, rqlQuery);
  }, [queryClient, rqlQuery]);
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
    { query: rqlQuery },
    {
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(
          lastPage,
          { query: rqlQuery },
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

  const handleLoadMore = useLoadMore({
    hasNextPage,
    isFetchingNextPage,
    isError,
    fetchNextPage,
    label: "audit logs",
  });

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
          <Navbar searchQuery={tableQuery.search} onExportCsv={onExportCsv} />
          <DataTable.Toolbar />
          <Flex className={styles["table-content-container"]}>
            <DataTable.VirtualizedContent
              classNames={{
                root: styles["table-wrapper"],
                table: tableClassName,
                header: styles["table-header"],
              }}
              emptyState={<NoAuditLogs />}
              rowHeight={60}
              groupHeaderHeight={48}
            />
            {sidePanelOpen && (
              <SidePanelDetails
                {...selectedAuditLog}
                onClose={() => setSidePanelOpen(false)}
                onNavigate={onNavigate}
              />
            )}
          </Flex>
        </Flex>
      </DataTable>
    </>
  );
}
