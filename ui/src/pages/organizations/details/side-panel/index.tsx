import { Avatar, SidePanel } from "@raystack/apsara/v1";
import { V1Beta1Organization, V1Beta1BillingAccount } from "~/api/frontier";
import { OrganizationDetailsSection } from "./org-details-section";
import { KYCDetailsSection } from "./kyc-section";
import { PlanDetailsSection } from "./plan-details-section";
import { TokensDetailsSection } from "./tokens-details-section";
import { useEffect, useState } from "react";
import { api } from "~/api";

export const SUBSCRIPTION_STATES = {
  active: "Active",
  past_due: "Past due",
  trialing: "Trialing",
  canceled: "Canceled",
  "": "NA",
} as const;

interface SidePanelProps {
  organization: V1Beta1Organization;
}

export function OrgSidePanel({ organization }: SidePanelProps) {
  const [isBillingAccountLoading, setIsBillingAccountLoading] = useState(true);
  const [billingAccount, setBillingAccount] = useState<V1Beta1BillingAccount>();

  useEffect(() => {
    async function fetchBillingAccount(orgId: string) {
      try {
        setIsBillingAccountLoading(true);
        const resp = await api?.frontierServiceListBillingAccounts(orgId);
        const newBillingAccount = resp.data?.billing_accounts?.[0];
        setBillingAccount(newBillingAccount);
      } catch (error) {
        console.error("Error fetching billing account:", error);
      } finally {
        setIsBillingAccountLoading(false);
      }
    }

    if (organization?.id) {
      fetchBillingAccount(organization.id);
    }
  }, [organization?.id]);

  return (
    <SidePanel data-test-id="admin-ui-sidepanel">
      <SidePanel.Header
        title={organization?.title || "Organization"}
        icon={<Avatar fallback={organization?.title?.[0]} />}
      />
      <SidePanel.Section>
        <OrganizationDetailsSection organization={organization} />
      </SidePanel.Section>
      <SidePanel.Section>
        <KYCDetailsSection organizationId={organization.id || ""} />
      </SidePanel.Section>
      <SidePanel.Section>
        <PlanDetailsSection
          organizationId={organization.id || ""}
          billingAccountId={billingAccount?.id || ""}
          isLoading={isBillingAccountLoading}
        />
      </SidePanel.Section>
      <SidePanel.Section>
        <TokensDetailsSection organizationId={organization.id || ""} />
      </SidePanel.Section>
    </SidePanel>
  );
}
