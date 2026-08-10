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
 * The organization's members keyed by id.
 *
 * This is the full, unpaginated member list, so it is fetched by the views
 * that need it rather than for every organization page. react-query dedupes
 * the request between callers sharing an org id.
 *
 * Pass `undefined`/empty to disable the query.
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
