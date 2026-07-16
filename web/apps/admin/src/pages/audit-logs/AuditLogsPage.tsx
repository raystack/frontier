import { AuditLogsView } from "@raystack/frontier/admin";
import { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { clients } from "~/connect/clients";
import { exportCsvFromStream } from "~/utils/helper";
import type { RQLExportRequest, RQLRequest } from "@raystack/proton/frontier";

const adminClient = clients.admin({ useBinary: true });

export default function AuditLogsPage() {
  const navigate = useNavigate();
  const onExportCsv = useCallback(async (query: RQLRequest) => {
    await exportCsvFromStream(
      adminClient.exportAuditRecords,
      { query: query as unknown as RQLExportRequest },
      "audit-logs.csv",
    );
  }, []);
  const onNavigate = useCallback(
    // Forward router state (e.g. the org id) to the destination page.
    (path: string, state?: { orgId?: string }) =>
      navigate(path, state ? { state } : undefined),
    [navigate],
  );

  return (
    <AuditLogsView onExportCsv={onExportCsv} onNavigate={onNavigate} />
  );
}
