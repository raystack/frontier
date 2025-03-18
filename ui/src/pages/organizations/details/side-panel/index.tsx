import { Avatar, Flex, List, Text, SidePanel } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import { V1Beta1Organization } from "~/api/frontier";
import dayjs from "dayjs";
import { useEffect, useState } from "react";
import { api } from "~/api";
import Skeleton from "react-loading-skeleton";
import { V1Beta1Subscription } from "@raystack/frontier/api-client";
import { V1Beta1Plan } from "@raystack/frontier";
import { OrganizationDetailsSection } from "./org-details-section";
import { KYCDetailsSection } from "./kyc-section";

export const SUBSCRIPTION_STATES = {
  active: "Active",
  past_due: "Past due",
  trialing: "Trialing",
  canceled: "Canceled",
  "": "NA",
} as const;

const BillingDetails = ({ organizationId }: { organizationId: string }) => {
  const [subscription, setSubscription] = useState<V1Beta1Subscription>();
  const [plan, setPlan] = useState<V1Beta1Plan>();

  useEffect(() => {
    async function fetchBillingDetails(id: string) {
      const subResponse = await api?.frontierServiceListSubscriptions2(id);
      const subscriptions = subResponse?.data?.subscriptions || [];
      const sub = subscriptions.find(
        (sub) => sub.state === "active" || sub.state === "trialing",
      );
      if (sub && sub.plan_id) {
        setSubscription(sub);
        const planResponse = await api?.frontierServiceGetPlan(sub.plan_id);
        const plan = planResponse?.data?.plan;
        setPlan(plan);
      }
    }
    if (organizationId) {
      fetchBillingDetails(organizationId);
    }
  }, [organizationId]);

  return (
    <List.Root>
      <List.Header>Plan details</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Name
        </List.Label>
        <List.Value>{plan?.title || "N/A"}</List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Started from
        </List.Label>
        <List.Value>
          {dayjs(subscription?.current_period_start_at).format("DD MMM YYYY")}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Ends on
        </List.Label>
        <List.Value>
          {dayjs(subscription?.current_period_end_at).format("DD MMM YYYY")}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Payment mode
        </List.Label>
        <List.Value>Prepaid</List.Value>
      </List.Item>
    </List.Root>
  );
};

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
        <BillingDetails organizationId={organization.id || ""} />
      </SidePanel.Section>
    </SidePanel>
  );
}
