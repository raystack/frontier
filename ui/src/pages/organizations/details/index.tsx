import {
  V1Beta1BillingAccount,
  V1Beta1Organization,
  V1Beta1Role,
} from "~/api/frontier";
import { useCallback, useEffect, useState } from "react";
import { api } from "~/api";
import { Outlet, useParams } from "react-router-dom";

import { OrganizationDetailsLayout } from "./layout";
import { ORG_NAMESPACE } from "./types";
import { OrganizationContext } from "./contexts/organization-context";

export const OrganizationDetails = () => {
  const [orgRoles, setOrgRoles] = useState<V1Beta1Role[]>([]);
  const [isOrgRolesLoading, setIsOrgRolesLoading] = useState(true);

  const [tokenBalance, setTokenBalance] = useState("0");
  const [isTokenBalanceLoading, setIsTokenBalanceLoading] = useState(false);

  const [isBillingAccountLoading, setIsBillingAccountLoading] = useState(true);
  const [billingAccount, setBillingAccount] = useState<V1Beta1BillingAccount>();

  const [organization, setOrganization] = useState<V1Beta1Organization>();
  const [isOrganizationLoading, setIsOrganizationLoading] = useState(true);

  const [isSearchVisible, setIsSearchVisible] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

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

  const fetchOrgTokenBalance = useCallback(
    async (orgId: string, billingAccountId: string) => {
      try {
        setIsTokenBalanceLoading(true);
        const resp = await api.frontierServiceGetBillingBalance(
          orgId,
          billingAccountId,
        );
        const newBalance = resp.data.balance?.amount || "0";
        setTokenBalance(newBalance);
      } catch (error) {
        console.error("Error fetching organization token balance:", error);
      } finally {
        setIsTokenBalanceLoading(false);
      }
    },
    [],
  );

  const fetchBillingAccount = useCallback(
    async (orgId: string) => {
      try {
        setIsBillingAccountLoading(true);
        const resp = await api?.frontierServiceListBillingAccounts(orgId);
        const newBillingAccount = resp.data?.billing_accounts?.[0];
        setBillingAccount(newBillingAccount);
        fetchOrgTokenBalance(orgId, newBillingAccount?.id || "");
      } catch (error) {
        console.error("Error fetching billing account:", error);
      } finally {
        setIsBillingAccountLoading(false);
      }
    },
    [fetchOrgTokenBalance],
  );

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
      fetchBillingAccount(organizationId);
    }
  }, [organizationId, fetchBillingAccount]);

  const isLoading =
    isOrganizationLoading ||
    isOrgRolesLoading ||
    isBillingAccountLoading ||
    isTokenBalanceLoading;

  return (
    <OrganizationContext.Provider
      value={{
        organization,
        roles: orgRoles,
        billingAccount,
        tokenBalance: tokenBalance,
        fetchTokenBalance: fetchOrgTokenBalance,
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
