import {
  OrganizationList,
} from "@raystack/frontier/admin";
import { useCallback, useContext, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { AppContext } from "~/contexts/App";
import { clients } from "~/connect/clients";
import { exportCsvFromStream } from "~/utils/helper";

const adminClient = clients.admin({ useBinary: true });

async function loadCountries(): Promise<string[]> {
  const data = await import("~/assets/data/countries.json");
  return (data.default as { name: string }[]).map((c) => c.name);
}

export function OrganizationListPage() {
  const navigate = useNavigate();
  const { config } = useContext(AppContext);
  const [countries, setCountries] = useState<string[]>([]);

  useEffect(() => {
    loadCountries().then(setCountries);
  }, []);

  const onNavigateToOrg = useCallback(
    (id: string) => navigate(`/organizations/${id}`),
    [navigate],
  );

  const onExportCsv = useCallback(async () => {
    await exportCsvFromStream(
      adminClient.exportOrganizations,
      {},
      "organizations.csv",
    );
  }, []);

  return (
    <OrganizationList
      appName={config?.title}
      appUrl={config?.app_url}
      organizationTypes={config?.organization_types}
      countries={countries}
      onNavigateToOrg={onNavigateToOrg}
      onExportCsv={onExportCsv}
    />
  );
}
