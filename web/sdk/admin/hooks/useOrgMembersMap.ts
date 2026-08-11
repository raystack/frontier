import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries, type User } from "@raystack/proton/frontier";
import type { ListOrganizationUsersResponse } from "@raystack/proton/frontier";

/* Module scope keeps the identity stable, so react-query can memoize it. */
const toMembersMap = (data?: ListOrganizationUsersResponse) =>
  (data?.users || []).reduce(
    (acc, user) => {
      acc[user.id || ""] = user;
      return acc;
    },
    {} as Record<string, User>,
  );

/**
 * The organization's members keyed by id — the full, unpaginated list, so it
 * is fetched by the views that need it rather than for every org page.
 * react-query dedupes it between callers. Pass empty to disable.
 */
export const useOrgMembersMap = (orgId?: string) =>
  useQuery(
    FrontierServiceQueries.listOrganizationUsers,
    { id: orgId || "" },
    {
      enabled: !!orgId,
      select: toMembersMap,
    },
  );
