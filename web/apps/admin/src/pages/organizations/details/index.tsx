import {
  OrganizationDetails,
} from "@raystack/frontier/admin";
import { useCallback, useContext, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { AppContext } from "~/contexts/App";
import { clients } from "~/connect/clients";
import { exportCsvFromStream } from "~/utils/helper";

const adminClient = clients.admin({ useBinary: true });

async function loadCountries(): Promise<string[]> {
  const data = await import("~/assets/data/countries.json");
  return (data.default as { name: string }[]).map((c) => c.name);
}

export function OrganizationDetailsPage() {
  const { organizationId } = useParams<{ organizationId: string }>();
  const { config } = useContext(AppContext);
  const [countries, setCountries] = useState<string[]>([]);

  useEffect(() => {
    loadCountries().then(setCountries);
  }, []);

  const onExportMembers = useCallback(async () => {
    if (!organizationId) return;
    await exportCsvFromStream(
      adminClient.exportOrganizationUsers,
      { id: organizationId },
      "organization-members.csv",
    );
  }, [organizationId]);

  const onExportProjects = useCallback(async () => {
    if (!organizationId) return;
    await exportCsvFromStream(
      adminClient.exportOrganizationProjects,
      { id: organizationId },
      "organization-projects.csv",
    );
  }, [organizationId]);

  const onExportTokens = useCallback(async () => {
    if (!organizationId) return;
    await exportCsvFromStream(
      adminClient.exportOrganizationTokens,
      { id: organizationId },
      "organization-tokens.csv",
    );
  }, [organizationId]);

  return (
    <OrganizationDetails
      organizationId={organizationId}
      appUrl={config?.app_url}
      tokenProductId={config?.token_product_id}
      countries={countries}
      organizationTypes={config?.organization_types}
      onExportMembers={onExportMembers}
      onExportProjects={onExportProjects}
      onExportTokens={onExportTokens}
    />
  );
}
