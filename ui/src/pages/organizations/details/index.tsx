import { V1Beta1Organization, V1Beta1Role } from "~/api/frontier";
import { useEffect, useState } from "react";
import { api } from "~/api";
import { Outlet, useParams } from "react-router-dom";

import { OrganizationDetailsLayout } from "./layout";
import { ORG_NAMESPACE } from "./types";
import { OrganizationContext } from "./contexts/organization-context";

export const OrganizationDetails = () => {
  const [orgRoles, setOrgRoles] = useState<V1Beta1Role[]>([]);
  const [isOrgRolesLoading, setIsOrgRolesLoading] = useState(true);
  const [isSearchVisible, setIsSearchVisible] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const [organization, setOrganization] = useState<V1Beta1Organization>();
  const [isOrganizationLoading, setIsOrganizationLoading] = useState(true);
  const { organizationId } = useParams();

  async function fetchRoles(orgId: string) {
    try {
      setIsOrgRolesLoading(true);
      const [defaultRolesResponse, organizationRolesResponse] =
        await Promise.all([
          api?.frontierServiceListRoles({
            scopes: [ORG_NAMESPACE],
          }),
          api?.frontierServiceListOrganizationRoles(orgId, {
            scopes: [ORG_NAMESPACE],
          }),
        ]);
      const defaultRoles = defaultRolesResponse.data?.roles || [];
      const organizationRoles = organizationRolesResponse.data?.roles || [];
      const roles = [...defaultRoles, ...organizationRoles];
      setOrgRoles(roles);
    } catch (error) {
      console.error(error);
    } finally {
      setIsOrgRolesLoading(false);
    }
  }

  async function fetchOrganization(id: string) {
    try {
      setIsOrganizationLoading(true);
      const response = await api?.frontierServiceGetOrganization(id);
      const org = response.data?.organization;
      setOrganization(org);
    } catch (error) {
      console.error(error);
    } finally {
      setIsOrganizationLoading(false);
    }
  }

  useEffect(() => {
    if (organizationId) {
      fetchOrganization(organizationId);
      fetchRoles(organizationId);
    }
  }, [organizationId]);

  const isLoading = isOrganizationLoading || isOrgRolesLoading;

  return (
    <OrganizationContext.Provider
      value={{
        organization,
        roles: orgRoles,
        search: {
          isVisible: isSearchVisible,
          setVisibility: setIsSearchVisible,
          query: searchQuery,
          onChange: setSearchQuery,
        },
      }}
    >
      <OrganizationDetailsLayout
        organization={organization}
        isLoading={isLoading}
      >
        {organization?.id ? (
          <Outlet
            context={{
              organizationId: organization?.id,
              fetchOrganization,
              organization,
            }}
          />
        ) : null}
      </OrganizationDetailsLayout>
    </OrganizationContext.Provider>
  );
};
