import { V1Beta1Subscription, V1Beta1Plan } from "~/api/frontier";
import { useEffect, useState } from "react";
import { api } from "~/api";
import { List, Text, Flex } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import dayjs from "dayjs";
import { CalendarIcon } from "@radix-ui/react-icons";

export const PlanDetailsSection = ({
  organizationId,
}: {
  organizationId: string;
}) => {
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
        <List.Value>
          <Text>{plan?.title || "Standard"}</Text>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Started from
        </List.Label>
        <List.Value>
          {subscription?.current_period_start_at ? (
            <Flex gap={3}>
              <CalendarIcon />
              <Text>
                {dayjs(subscription?.current_period_start_at).format(
                  "DD MMM YYYY",
                )}
              </Text>
            </Flex>
          ) : (
            <Text>-</Text>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Ends on
        </List.Label>
        <List.Value>
          {subscription?.current_period_end_at ? (
            <Flex gap={3}>
              <CalendarIcon />
              <Text>
                {dayjs(subscription?.current_period_end_at).format(
                  "DD MMM YYYY",
                )}
              </Text>
            </Flex>
          ) : (
            <Text>-</Text>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Payment mode
        </List.Label>
        <List.Value>
          <Text>Prepaid</Text>
        </List.Value>
      </List.Item>
    </List.Root>
  );
};
