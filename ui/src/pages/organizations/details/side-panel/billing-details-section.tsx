import { Link, List, Text } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import { V1Beta1BillingAccount } from "~/api/frontier";
import { converBillingAddressToString } from "~/utils/helper";
import Skeleton from "react-loading-skeleton";

interface BillingDetailsSectionProps {
  organizationId: string;
  billingAccountId: string;
  isLoading: boolean;
  billingAccount?: V1Beta1BillingAccount;
}

export const BillingDetailsSection = ({
  isLoading,
  billingAccount,
}: BillingDetailsSectionProps) => {
  const isDataLoading = isLoading;

  const stripeLink = billingAccount?.provider_id
    ? "https://dashboard.stripe.com/customers/" + billingAccount?.provider_id
    : "";

  return (
    <List.Root>
      <List.Header>Billing</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Name
        </List.Label>
        <List.Value className={styles["side-panel-section-item-value"]}>
          {isDataLoading ? (
            <Skeleton />
          ) : (
            <Text>{billingAccount?.name || "N/A"}</Text>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Email
        </List.Label>
        <List.Value className={styles["side-panel-section-item-value"]}>
          {isDataLoading ? (
            <Skeleton />
          ) : (
            <Text>{billingAccount?.email || "N/A"}</Text>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Address
        </List.Label>
        <List.Value className={styles["side-panel-section-item-value"]}>
          {isDataLoading ? (
            <Skeleton />
          ) : (
            <Text>
              {billingAccount?.address
                ? converBillingAddressToString(billingAccount.address)
                : "N/A"}
            </Text>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Next billing date
        </List.Label>
        <List.Value className={styles["side-panel-section-item-value"]}>
          {isDataLoading ? <Skeleton /> : <Text>-</Text>}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Amount
        </List.Label>
        <List.Value className={styles["side-panel-section-item-value"]}>
          {isDataLoading ? <Skeleton /> : <Text>-</Text>}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Profile
        </List.Label>
        <List.Value className={styles["side-panel-section-item-value"]}>
          {isDataLoading ? (
            <Skeleton />
          ) : stripeLink ? (
            <Link
              href={stripeLink}
              data-test-id="stripe-dashboard-link"
              target="_blank"
            >
              Stripe
            </Link>
          ) : (
            <Text>N/A</Text>
          )}
        </List.Value>
      </List.Item>
    </List.Root>
  );
};
