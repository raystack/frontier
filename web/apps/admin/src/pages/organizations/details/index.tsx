import { OrganizationDetailsView, useAdminPaths } from '@raystack/frontier/admin';
import { useCallback, useContext, useEffect, useState } from 'react';
import { useLocation, useNavigate, useParams, Outlet, Navigate } from 'react-router-dom';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { AppContext } from '~/contexts/App';
import { clients } from '~/connect/clients';
import { exportCsvFromStream } from '~/utils/helper';
import LoadingState from '~/components/states/Loading';

const adminClient = clients.admin({ useBinary: true });

async function loadCountries(): Promise<string[]> {
  const data = await import("~/assets/data/countries.json");
  return (data.default as { name: string }[]).map((c) => c.name);
}

export default function OrganizationDetailsPage() {
  /*
   * RPCs run by id:
   * - in-app nav passes the id via router state (fast path, works for any org)
   * - cold load (refresh/bookmark) → resolve the URL id/slug instead
   * - a disabled org's slug can't resolve → redirect to the list
   */
  const { organizationId: urlParam } = useParams<{ organizationId: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const paths = useAdminPaths();
  const { config } = useContext(AppContext);
  const [countries, setCountries] = useState<string[]>([]);

  const [stateOrgId, setStateOrgId] = useState<string | undefined>(
    (location.state as { orgId?: string } | null)?.orgId,
  );

  // New id on org switch; keep the current one on tab switches.
  useEffect(() => {
    const next = (location.state as { orgId?: string } | null)?.orgId;
    if (next && next !== stateOrgId) setStateOrgId(next);
  }, [location.state, stateOrgId]);

  // Cold-load fallback: resolve from the URL only when state has no id.
  const needsResolve = !stateOrgId && !!urlParam;
  const { data: resolvedId, error: resolveError } = useQuery(
    FrontierServiceQueries.getOrganization,
    { id: urlParam || "" },
    { enabled: needsResolve, select: (data) => data?.organization?.id },
  );

  const orgId = stateOrgId || resolvedId;

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

  // No id, or the URL didn't resolve → list.
  if (!urlParam || resolveError) {
    return <Navigate to={`/${paths.organizations}`} replace />;
  }

  // Still resolving the cold-load id — show a loader, don't flash a redirect.
  if (!orgId) {
    return <LoadingState />;
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
