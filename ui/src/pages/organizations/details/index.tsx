import { useCallback, useEffect, useState } from "react";
import { useQuery } from "@connectrpc/connect-query";
import { Outlet, useParams } from "react-router-dom";

import { OrganizationDetailsLayout } from "./layout";
import { ORG_NAMESPACE } from "./types";
import { OrganizationContext } from "./contexts/organization-context";
import {
  FrontierServiceQueries,
  type Organization,
  type User,
} from "@raystack/proton/frontier";
import { queryClient } from "~/contexts/ConnectProvider";

export const OrganizationDetails = () => {
  const [isSearchVisible, setIsSearchVisible] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  const { organizationId } = useParams();

  // Use Connect RPC for fetching organization
  const {
    data: organization,
    isLoading: isOrganizationLoading,
    error: organizationError,
  } = useQuery(
    FrontierServiceQueries.getOrganization,
    { id: organizationId },
    {
      enabled: !!organizationId,
      select: (data) => data?.organization,
    },
  );

  const getOrganizationQueryKey = [
    FrontierServiceQueries.getOrganization,
    { id: organizationId },
  ];

  async function updateOrganization(org: Organization) {
    queryClient.setQueryData(getOrganizationQueryKey, { organization: org });
  }

  // Fetch KYC details
  const {
    data: kycDetails,
    isLoading: isKYCLoading,
    error: kycError,
  } = useQuery(
    FrontierServiceQueries.getOrganizationKyc,
    { orgId: organizationId || "" },
    {
      enabled: !!organizationId,
      select: (data) => data?.organizationKyc,
    },
  );

  function updateKYCDetails(kyc: typeof kycDetails) {
    if (!organizationId) return;
    queryClient.setQueryData(
      [FrontierServiceQueries.getOrganizationKyc, { orgId: organizationId }],
      { organizationKyc: kyc },
    );
  }

  // Fetch default roles
  const {
    data: defaultRoles = [],
    isLoading: isDefaultRolesLoading,
    error: defaultRolesError,
  } = useQuery(
    FrontierServiceQueries.listRoles,
    { scopes: [ORG_NAMESPACE] },
    {
      enabled: !!organizationId,
      select: (data) => data?.roles || [],
    },
  );

  // Fetch organization-specific roles
  const {
    data: organizationRoles = [],
    isLoading: isOrgRolesLoading,
    error: orgRolesError,
  } = useQuery(
    FrontierServiceQueries.listOrganizationRoles,
    { orgId: organizationId || "", scopes: [ORG_NAMESPACE] },
    {
      enabled: !!organizationId,
      select: (data) => data?.roles || [],
    },
  );

  const roles = [...defaultRoles, ...organizationRoles];

  // Fetch organization members
  const {
    data: orgMembersMap = {},
    isLoading: isOrgMembersMapLoading,
    error: orgMembersError,
  } = useQuery(
    FrontierServiceQueries.listOrganizationUsers,
    { id: organizationId || "" },
    {
      enabled: !!organizationId,
      select: (data) => {
        const users = data?.users || [];
        return users.reduce(
          (acc, user) => {
            const id = user.id || "";
            acc[id] = user;
            return acc;
          },
          {} as Record<string, User>,
        );
      },
    },
  );

  // Fetch billing accounts list
  const {
    data: firstBillingAccountId = "",
    error: billingAccountsError,
  } = useQuery(
    FrontierServiceQueries.listBillingAccounts,
    { orgId: organizationId || "" },
    {
      enabled: !!organizationId,
      select: (data) => data?.billingAccounts?.[0]?.id || "",
    },
  );

  // Fetch billing account details
  const {
    data: billingAccountData,
    isLoading: isBillingAccountLoading,
    error: billingAccountError,
    refetch: fetchBillingAccountDetails,
  } = useQuery(
    FrontierServiceQueries.getBillingAccount,
    {
      orgId: organizationId || "",
      id: firstBillingAccountId,
      withBillingDetails: true,
    },
    {
      enabled: !!organizationId && !!firstBillingAccountId,
      select: (data) => ({
        billingAccount: data?.billingAccount,
        billingAccountDetails: data?.billingDetails,
      }),
    },
  );

  const billingAccount = billingAccountData?.billingAccount;
  const billingAccountDetails = billingAccountData?.billingAccountDetails;

  // Fetch billing balance
  const {
    data: tokenBalance = "0",
    isLoading: isTokenBalanceLoading,
    error: tokenBalanceError,
    refetch: fetchTokenBalance
  } = useQuery(
    FrontierServiceQueries.getBillingBalance,
    {
      orgId: organizationId || "",
      id: firstBillingAccountId,
    },
    {
      enabled: !!organizationId && !!firstBillingAccountId,
      select: (data) => String(data?.balance?.amount || "0"),
    },
  );

  // Error handling
  useEffect(() => {
    if (organizationError) {
      console.error("Failed to fetch organization:", organizationError);
    }
    if (kycError) {
      console.error("Failed to fetch KYC details:", kycError);
    }
    if (defaultRolesError) {
      console.error("Failed to fetch default roles:", defaultRolesError);
    }
    if (orgRolesError) {
      console.error("Failed to fetch organization roles:", orgRolesError);
    }
    if (orgMembersError) {
      console.error("Failed to fetch organization members:", orgMembersError);
    }
    if (billingAccountsError) {
      console.error("Failed to fetch billing accounts:", billingAccountsError);
    }
    if (billingAccountError) {
      console.error("Failed to fetch billing account details:", billingAccountError);
    }
    if (tokenBalanceError) {
      console.error("Failed to fetch token balance:", tokenBalanceError);
    }
  }, [
    organizationError,
    kycError,
    defaultRolesError,
    orgRolesError,
    orgMembersError,
    billingAccountsError,
    billingAccountError,
    tokenBalanceError,
  ]);

  const isLoading =
    isOrganizationLoading ||
    isDefaultRolesLoading ||
    isOrgRolesLoading ||
    isBillingAccountLoading;
  return (
    <OrganizationContext.Provider
      value={{
        organization: organization,
        updateOrganization,
        roles,
        billingAccount,
        billingAccountDetails,
        isBillingAccountLoading,
        fetchBillingAccountDetails,
        tokenBalance,
        isTokenBalanceLoading,
        fetchTokenBalance,
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
              organization,
            }}
          />
        ) : null}
      </OrganizationDetailsLayout>
    </OrganizationContext.Provider>
  );
};
