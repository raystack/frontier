import { Flex, Link, List, Text } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import { V1Beta1BillingAccount, V1Beta1Invoice } from "~/api/frontier";
import { converBillingAddressToString } from "~/utils/helper";
import Skeleton from "react-loading-skeleton";
import { useEffect, useState } from "react";
import { api } from "~/api";
import dayjs from "dayjs";
import { CalendarIcon } from "@radix-ui/react-icons";
import { Amount } from "@raystack/frontier/react";

interface BillingDetailsSectionProps {
  organizationId: string;
  billingAccountId: string;
  isLoading: boolean;
  billingAccount?: V1Beta1BillingAccount;
}

export const BillingDetailsSection = ({
  isLoading,
  billingAccount,
  organizationId,
}: BillingDetailsSectionProps) => {
  const [upcomingInvoice, setUpcomingInvoice] = useState<V1Beta1Invoice>();
  const [isUpcomingInvoiceLoading, setIsUpcomingInvoiceLoading] =
    useState(false);

  useEffect(() => {
    async function getUpcomingInvoice(orgId: string, billingId: string) {
      setIsUpcomingInvoiceLoading(true);
      try {
        const resp = await api?.frontierServiceGetUpcomingInvoice(
          orgId,
          billingId,
        );
        const invoice = resp?.data?.invoice;
        if (invoice && invoice.state) {
          setUpcomingInvoice(invoice);
        }
      } catch (err: any) {
        console.error(err);
      } finally {
        setIsUpcomingInvoiceLoading(false);
      }
    }

    if (organizationId && billingAccount?.id) {
      getUpcomingInvoice(organizationId, billingAccount.id);
    }
  }, [organizationId, billingAccount]);

  const isDataLoading = isLoading || isUpcomingInvoiceLoading;

  const due_date = upcomingInvoice?.due_date || upcomingInvoice?.period_end_at;

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
          {isDataLoading ? (
            <Skeleton />
          ) : due_date ? (
            <Flex gap={3}>
              <CalendarIcon />
              <Text>{dayjs(due_date).format("DD MMM YYYY")}</Text>
            </Flex>
          ) : (
            <Text>-</Text>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Amount
        </List.Label>
        <List.Value className={styles["side-panel-section-item-value"]}>
          {isDataLoading ? (
            <Skeleton />
          ) : upcomingInvoice?.amount ? (
            <Amount
              currency={upcomingInvoice?.currency}
              value={Number(upcomingInvoice?.amount)}
            />
          ) : (
            <Text>-</Text>
          )}
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
