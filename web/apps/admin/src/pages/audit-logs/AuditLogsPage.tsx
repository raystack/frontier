import { AuditLogsView } from "@raystack/frontier/admin";
import { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { clients } from "~/connect/clients";
import { exportCsvFromStream } from "~/utils/helper";
import type { RQLExportRequest, RQLRequest } from "@raystack/proton/frontier";

const adminClient = clients.admin({ useBinary: true });

export function AuditLogsPage() {
  const navigate = useNavigate();
  const onExportCsv = useCallback(async (query: RQLRequest) => {
    await exportCsvFromStream(
      adminClient.exportAuditRecords,
      { query: query as unknown as RQLExportRequest },
      "audit-logs.csv",
    );
  }, []);
  const onNavigate = useCallback((path: string) => navigate(path), [navigate]);

  return (
    <AuditLogsView onExportCsv={onExportCsv} onNavigate={onNavigate} />
  );
}
