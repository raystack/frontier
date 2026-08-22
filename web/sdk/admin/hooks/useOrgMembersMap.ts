import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries, type User } from "@raystack/proton/frontier";
import type { ListOrganizationUsersResponse } from "@raystack/proton/frontier";

// Stable identity so react-query memoizes the select.
const toMembersMap = (data?: ListOrganizationUsersResponse) =>
  (data?.users || []).reduce(
    (acc, user) => {
      acc[user.id || ""] = user;
      return acc;
    },
    {} as Record<string, User>,
  );

/** Org members keyed by id. Deduped across callers; empty orgId disables. */
export const useOrgMembersMap = (orgId?: string) =>
  useQuery(
    FrontierServiceQueries.listOrganizationUsers,
    { id: orgId || "" },
    {
      enabled: !!orgId,
      select: toMembersMap,
    },
  );
