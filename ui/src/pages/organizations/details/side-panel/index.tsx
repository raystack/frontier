import { Avatar, SidePanel } from "@raystack/apsara/v1";
import { V1Beta1Organization } from "~/api/frontier";
import { OrganizationDetailsSection } from "./org-details-section";
import { KYCDetailsSection } from "./kyc-section";
import { PlanDetailsSection } from "./plan-details-section";
import { TokensDetailsSection } from "./tokens-details-section";
import { BillingDetailsSection } from "./billing-details-section";

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
        <PlanDetailsSection />
      </SidePanel.Section>
      <SidePanel.Section>
        <TokensDetailsSection />
      </SidePanel.Section>
      <SidePanel.Section>
        <BillingDetailsSection />
      </SidePanel.Section>
    </SidePanel>
  );
}
