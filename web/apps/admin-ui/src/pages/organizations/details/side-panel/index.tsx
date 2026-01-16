import { Avatar, getAvatarColor, SidePanel } from "@raystack/apsara";
import { OrganizationDetailsSection } from "./org-details-section";
import { KYCDetailsSection } from "./kyc-section";
import { PlanDetailsSection } from "./plan-details-section";
import { TokensDetailsSection } from "./tokens-details-section";
import { BillingDetailsSection } from "./billing-details-section";
import type { Organization } from "@raystack/proton/frontier";

export const SUBSCRIPTION_STATES = {
  active: "Active",
  past_due: "Past due",
  trialing: "Trialing",
  canceled: "Canceled",
  "": "NA",
} as const;

interface SidePanelProps {
  organization: Organization;
}

export function OrgSidePanel({ organization }: SidePanelProps) {
  const avatarColor = getAvatarColor(organization?.id || "");
  return (
    <SidePanel data-test-id="admin-ui-sidepanel">
      <SidePanel.Header
        title={organization?.title || "Organization"}
        icon={
          <Avatar fallback={organization?.title?.[0]} color={avatarColor} />
        }
      />
      <SidePanel.Section>
        <OrganizationDetailsSection organization={organization} />
      </SidePanel.Section>
      <SidePanel.Section>
        <KYCDetailsSection />
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
