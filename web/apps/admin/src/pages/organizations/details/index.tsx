import { OrganizationDetailsView, useAdminPaths } from '@raystack/frontier/admin';
import { useCallback, useContext, useEffect, useState } from 'react';
import { useLocation, useNavigate, Outlet, Navigate } from 'react-router-dom';
import { AppContext } from '~/contexts/App';
import { clients } from '~/connect/clients';
import { exportCsvFromStream } from '~/utils/helper';

const adminClient = clients.admin({ useBinary: true });

async function loadCountries(): Promise<string[]> {
  const data = await import("~/assets/data/countries.json");
  return (data.default as { name: string }[]).map((c) => c.name);
}

export default function OrganizationDetailsPage() {
  // URL shows the slug, but RPCs need the id (a disabled org's slug won't
  // resolve). The id comes in via router state and we hold it across tab
  // switches (which clear that state). No id (direct load/refresh) → go to list.
  const location = useLocation();
  const navigate = useNavigate();
  const paths = useAdminPaths();
  const { config } = useContext(AppContext);
  const [countries, setCountries] = useState<string[]>([]);
  const [orgId, setOrgId] = useState<string | undefined>(
    (location.state as { orgId?: string } | null)?.orgId,
  );

  // Pick up a new id when switching orgs; keep the current one on tab switches.
  useEffect(() => {
    const stateOrgId = (location.state as { orgId?: string } | null)?.orgId;
    if (stateOrgId && stateOrgId !== orgId) setOrgId(stateOrgId);
  }, [location.state, orgId]);

  useEffect(() => {
    loadCountries().then(setCountries);
  }, []);

  const onExportMembers = useCallback(async () => {
    if (!orgId) return;
    await exportCsvFromStream(
      adminClient.exportOrganizationUsers,
      { id: orgId },
      "organization-members.csv",
    );
  }, [orgId]);

  const onExportProjects = useCallback(async () => {
    if (!orgId) return;
    await exportCsvFromStream(
      adminClient.exportOrganizationProjects,
      { id: orgId },
      "organization-projects.csv",
    );
  }, [orgId]);

  const onExportTokens = useCallback(async () => {
    if (!orgId) return;
    await exportCsvFromStream(
      adminClient.exportOrganizationTokens,
      { id: orgId },
      "organization-tokens.csv",
    );
  }, [orgId]);

  if (!orgId) {
    return <Navigate to={`/${paths.organizations}`} replace />;
  }

  return (
    <OrganizationDetailsView
      organizationId={orgId}
      appUrl={config?.app_url}
      tokenProductId={config?.token_product_id}
      countries={countries}
      organizationTypes={config?.organization_types}
      onExportMembers={onExportMembers}
      onExportProjects={onExportProjects}
      onExportTokens={onExportTokens}
      currentPath={location.pathname}
      onNavigate={navigate}
    >
      <Outlet />
    </OrganizationDetailsView>
  );
}
