import { CopyButton, Flex, Link, List, Text } from "@raystack/apsara";
import styles from "./side-panel.module.css";
import { converBillingAddressToString } from "../../../../utils/helper";
import Skeleton from "react-loading-skeleton";
import { useContext, useEffect } from "react";
import { CalendarIcon } from "@radix-ui/react-icons";
import { Amount } from "@raystack/apsara";
import { OrganizationContext } from "../contexts/organization-context";
import { useQuery } from "@connectrpc/connect-query";
import { FrontierServiceQueries, GetUpcomingInvoiceRequestSchema } from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { timestampToDayjs } from "../../../../utils/connect-timestamp";

export const BillingDetailsSection = () => {
  const { billingAccount, organization } = useContext(OrganizationContext);

  const organizationId = organization?.id || "";
  const billingAccountId = billingAccount?.id || "";

  const { data: upcomingInvoice, isLoading, error } = useQuery(
    FrontierServiceQueries.getUpcomingInvoice,
    create(GetUpcomingInvoiceRequestSchema, {
      orgId: organizationId,
      billingId: billingAccountId,
    }),
    {
      enabled: !!organizationId && !!billingAccountId,
      select: (data) => data?.invoice,
    }
  );

  useEffect(() => {
    if (error) {
      console.error("Error fetching upcoming invoice:", error);
    }
  }, [error]);
  const due_date = upcomingInvoice?.dueDate || upcomingInvoice?.periodEndAt;

  const stripeLink = billingAccount?.providerId
    ? "https://dashboard.stripe.com/customers/" + billingAccount?.providerId
    : "";

  return (
    <List>
      <List.Header>Billing</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Name
        </List.Label>
        <List.Value className={styles["side-panel-section-item-value"]}>
          {isLoading ? (
            <Skeleton />
          ) : (
            <Text>{billingAccount?.name || "N/A"}</Text>
          )}
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Billing Account ID
        </List.Label>
        <List.Value>
          <Flex gap={3}>
            <CopyButton
              text={billingAccount?.id || ""}
              data-test-id="copy-button"
            />
            <Text className={styles["org-details-section-org-id"]}>
              {billingAccount?.id}
            </Text>
          </Flex>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Email
        </List.Label>
        <List.Value className={styles["side-panel-section-item-value"]}>
          {isLoading ? (
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
          {isLoading ? (
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
          {isLoading ? (
            <Skeleton />
          ) : timestampToDayjs(due_date) ? (
            <Flex gap={3}>
              <CalendarIcon />
              <Text>{timestampToDayjs(due_date)?.format("DD MMM YYYY")}</Text>
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
          {isLoading ? (
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
          {isLoading ? (
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
    </List>
  );
};
