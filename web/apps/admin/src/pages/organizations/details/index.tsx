import { OrganizationDetailsView, useAdminPaths } from '@raystack/frontier/admin';
import { useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { useLocation, useNavigate, useParams, Outlet, Navigate } from 'react-router-dom';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  AdminServiceQueries,
  RQLRequestSchema,
  RQLFilterSchema,
} from '@raystack/proton/frontier';
import { AppContext } from '~/contexts/App';
import { clients } from '~/connect/clients';
import { exportCsvFromStream } from '~/utils/helper';
import LoadingState from '~/components/states/Loading';

const adminClient = clients.admin({ useBinary: true });

/* Old bookmarks use the id; new URLs use the slug — tell them apart. */
const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

async function loadCountries(): Promise<string[]> {
  const data = await import("~/assets/data/countries.json");
  return (data.default as { name: string }[]).map((c) => c.name);
}

export default function OrganizationDetailsPage() {
  /*
   * View runs by org id:
   * - in-app nav: id via router state (fast path)
   * - cold load: resolve URL slug/id via admin API (returns disabled orgs too)
   * - unresolvable URL: redirect to list
   */
  const { organizationId: urlParam } = useParams<{ organizationId: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const paths = useAdminPaths();
  const { config } = useContext(AppContext);
  const [countries, setCountries] = useState<string[]>([]);

  const incomingOrgId = (location.state as { orgId?: string } | null)?.orgId;

  /*
   * Hold the id keyed to its URL, recomputed during render:
   * - a urlParam change invalidates a stale id, re-arming the resolve
   * - setState-during-render keeps children from seeing the old id (no flash,
   *   no stale org under a new URL)
   */
  const [held, setHeld] = useState<{ param?: string; orgId?: string }>({
    param: urlParam,
    orgId: incomingOrgId,
  });
  if (held.param !== urlParam || (incomingOrgId && incomingOrgId !== held.orgId)) {
    setHeld({ param: urlParam, orgId: incomingOrgId });
  }
  const stateOrgId = held.param === urlParam ? held.orgId : undefined;

  /*
   * Cold-load resolve (only when state has no id):
   * - UUID param is already the id
   * - slug → admin searchOrganizations, which returns disabled orgs too
   *   (getOrganization-by-slug does not)
   */
  const paramIsId = !!urlParam && UUID_RE.test(urlParam);
  const needsResolve = !stateOrgId && !!urlParam && !paramIsId;
  const searchReq = useMemo(
    () => ({
      query: create(RQLRequestSchema, {
        filters: needsResolve
          ? [
              create(RQLFilterSchema, {
                name: 'name',
                operator: 'eq',
                value: { case: 'stringValue', value: urlParam as string },
              }),
            ]
          : [],
        limit: 1,
      }),
    }),
    [needsResolve, urlParam],
  );
  const {
    data: searchedId,
    error: resolveError,
    isSuccess,
  } = useQuery(AdminServiceQueries.searchOrganizations, searchReq, {
    enabled: needsResolve,
    staleTime: 5 * 60 * 1000,
    select: (d) => d?.organizations?.[0]?.id,
  });

  const orgId = stateOrgId || (paramIsId ? urlParam : searchedId);
  const notFound = needsResolve && isSuccess && !searchedId;

  /* Carry the org id in router state so breadcrumb/tab navs skip the resolve. */
  const onNavigate = useCallback(
    (path: string, state?: { orgId?: string }) =>
      navigate(path, state ? { state } : undefined),
    [navigate],
  );

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

  /* No id, error, or empty result (also guards the infinite loader) → list. */
  if (!urlParam || resolveError || notFound) {
    return <Navigate to={`/${paths.organizations}`} replace />;
  }

  /* Still resolving — loader, not a redirect flash. */
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
      onNavigate={onNavigate}
    >
      <Outlet />
    </OrganizationDetailsView>
  );
}
