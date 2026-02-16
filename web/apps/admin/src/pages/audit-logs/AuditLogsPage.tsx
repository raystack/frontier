import { AuditLogsView } from "@raystack/frontier/admin";
import { useCallback } from "react";
import { clients } from "~/connect/clients";
import { exportCsvFromStream } from "~/utils/helper";
import type { RQLExportRequest, RQLRequest } from "@raystack/proton/frontier";

const adminClient = clients.admin({ useBinary: true });

export function AuditLogsPage() {
  const onExportCsv = useCallback(async (query: RQLRequest) => {
    await exportCsvFromStream(
      adminClient.exportAuditRecords,
      { query: query as unknown as RQLExportRequest },
      "audit-logs.csv",
    );
  }, []);

  return <AuditLogsView onExportCsv={onExportCsv} />;
}
