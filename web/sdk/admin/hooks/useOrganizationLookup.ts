import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";

/**
 * Fetches an organization by id — for display (its title) and to build a slug
 * link. Call it with the id: a disabled org's slug no longer resolves
 * server-side, so the id is the reliable key. react-query dedupes/caches repeat
 * lookups of the same id across callers.
 *
 * Pass `undefined`/empty to disable the query (e.g. when there's no org yet).
 */
export const useOrganizationLookup = (orgId?: string) =>
  useQuery(
    FrontierServiceQueries.getOrganization,
    { id: orgId || "" },
    {
      enabled: !!orgId,
      staleTime: 5 * 60 * 1000,
      select: (data) => data?.organization,
    },
  );
