import { useEffect, useMemo } from "react";
import { useQuery } from "@connectrpc/connect-query";
import { create } from "@bufbuild/protobuf";
import {
  FrontierServiceQueries,
  ListRolesRequestSchema,
  ListOrganizationRolesRequestSchema,
} from "@raystack/proton/frontier";
import { SCOPES, SHARED_QUERY_STALE_TIME } from "~/admin/utils/constants";

interface UseOrganizationRolesOptions {
  /** Skip both fetches while false. Defaults to true. */
  enabled?: boolean;
}

/*
  Roles assignable within an org: the platform's defaults plus the org's custom
  ones. Both halves are needed — a role id can come from either.
  - nothing writes roles, so repeat callers share one cached fetch
  - pass undefined/empty to skip the org-scoped half
*/
export const useOrganizationRoles = (
  orgId?: string,
  { enabled = true }: UseOrganizationRolesOptions = {},
) => {
  const {
    data: defaultRoles = [],
    isLoading: isDefaultRolesLoading,
    error: defaultRolesError,
  } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, { scopes: [SCOPES.ORG] }),
    {
      enabled,
      staleTime: SHARED_QUERY_STALE_TIME,
      select: (data) => data?.roles || [],
    },
  );

  const {
    data: organizationRoles = [],
    isLoading: isOrgRolesLoading,
    error: orgRolesError,
  } = useQuery(
    FrontierServiceQueries.listOrganizationRoles,
    create(ListOrganizationRolesRequestSchema, {
      orgId: orgId || "",
      scopes: [SCOPES.ORG],
    }),
    {
      enabled: enabled && !!orgId,
      staleTime: SHARED_QUERY_STALE_TIME,
      select: (data) => data?.roles || [],
    },
  );

  useEffect(() => {
    if (defaultRolesError) {
      console.error("Failed to fetch default roles:", defaultRolesError);
    }
    if (orgRolesError) {
      console.error("Failed to fetch organization roles:", orgRolesError);
    }
  }, [defaultRolesError, orgRolesError]);

  const roles = useMemo(
    () => [...defaultRoles, ...organizationRoles],
    [defaultRoles, organizationRoles],
  );

  const titleById = useMemo(
    () => new Map(roles.map((role) => [role.id, role.title || role.name])),
    [roles],
  );

  return {
    roles,
    titleById,
    isLoading: isDefaultRolesLoading || isOrgRolesLoading,
    error: defaultRolesError ?? orgRolesError,
  };
};
