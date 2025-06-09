import type {
  V1Beta1BillingAccount,
  V1Beta1BillingAccountDetails,
  V1Beta1Organization,
  V1Beta1OrganizationKyc,
  V1Beta1Role,
  V1Beta1User,
} from "~/api/frontier";
import { useCallback, useEffect, useState } from "react";
import { api } from "~/api";
import { Outlet, useParams } from "react-router-dom";

import { OrganizationDetailsLayout } from "./layout";
import { ORG_NAMESPACE } from "./types";
import { OrganizationContext } from "./contexts/organization-context";
import { AxiosError } from "axios";

export const OrganizationDetails = () => {
  const [orgRoles, setOrgRoles] = useState<V1Beta1Role[]>([]);
  const [isOrgRolesLoading, setIsOrgRolesLoading] = useState(true);

  const [tokenBalance, setTokenBalance] = useState("0");
  const [isTokenBalanceLoading, setIsTokenBalanceLoading] = useState(false);

  const [isBillingAccountLoading, setIsBillingAccountLoading] = useState(true);
  const [billingAccount, setBillingAccount] = useState<V1Beta1BillingAccount>();
  const [billingAccountDetails, setBillingAccountDetails] =
    useState<V1Beta1BillingAccountDetails>();

  const [organization, setOrganization] = useState<V1Beta1Organization>();
  const [isOrganizationLoading, setIsOrganizationLoading] = useState(true);

  const [isSearchVisible, setIsSearchVisible] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const [isOrgMembersMapLoading, setIsOrgMembersMapLoading] = useState(false);
  const [orgMembersMap, setOrgMembersMap] = useState<
    Record<string, V1Beta1User>
  >({});

  const [kycDetails, setKycDetails] = useState<
    V1Beta1OrganizationKyc | undefined
  >();
  const [isKYCLoading, setIsKYCLoading] = useState(true);
  const { organizationId } = useParams();

  async function fetchKYCDetails(id: string) {
    setIsKYCLoading(true);
    try {
      const response = await api?.frontierServiceGetOrganizationKyc(id);
      const kyc = response?.data?.organization_kyc;
      setKycDetails(kyc);
    } catch (error: unknown) {
      if (error instanceof AxiosError && error.response?.status === 404) {
        console.warn("KYC details not found");
      } else {
        console.error("Error fetching KYC details:", error);
      }
    } finally {
      setIsKYCLoading(false);
    }
  }

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

  async function fetchOrgMembers(orgId: string) {
    try {
      setIsOrgMembersMapLoading(true);
      const [orgUserResp] = await Promise.all([
        api?.frontierServiceListOrganizationUsers(orgId),
      ]);
      const orgUsers = orgUserResp.data?.users || [];
      const orgUsersMap = orgUsers.reduce(
        (acc, user) => {
          const id = user.id || "";
          acc[id] = user;
          return acc;
        },
        {} as Record<string, V1Beta1User>,
      );
      setOrgMembersMap(orgUsersMap);
    } catch (error) {
      console.error(error);
    } finally {
      setIsOrgMembersMapLoading(false);
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
        const listBillingResp =
          await api?.frontierServiceListBillingAccounts(orgId);
        const firstBillingAccount = listBillingResp.data?.billing_accounts?.[0];
        const getBillingResp = await api?.frontierServiceGetBillingAccount(
          orgId,
          firstBillingAccount?.id || "",
          { with_billing_details: true },
        );

        const newBillingAccount = getBillingResp.data?.billing_account;
        const newBillingAccountDetails = getBillingResp.data.billing_details;
        setBillingAccountDetails(newBillingAccountDetails);
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

  function updateKYCDetails(kycDetails: V1Beta1OrganizationKyc) {
    setKycDetails(kycDetails);
  }

  async function updateOrganization(org: V1Beta1Organization) {
    setOrganization(org);
  }

  useEffect(() => {
    if (organizationId) {
      fetchOrganization(organizationId);
      fetchRoles(organizationId);
      fetchBillingAccount(organizationId);
      fetchOrgMembers(organizationId);
      fetchKYCDetails(organizationId);
    }
  }, [organizationId, fetchBillingAccount]);

  const isLoading =
    isOrganizationLoading || isOrgRolesLoading || isBillingAccountLoading;
  return (
    <OrganizationContext.Provider
      value={{
        organization,
        updateOrganization,
        roles: orgRoles,
        billingAccount,
        billingAccountDetails,
        setBillingAccountDetails,
        tokenBalance: tokenBalance,
        isTokenBalanceLoading,
        fetchTokenBalance: fetchOrgTokenBalance,
        orgMembersMap,
        isOrgMembersMapLoading,
        updateKYCDetails,
        kycDetails,
        isKYCLoading,
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
