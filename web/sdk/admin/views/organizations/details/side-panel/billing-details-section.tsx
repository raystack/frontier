import { Amount, CopyButton, Flex, IconButton, Link, List, Text, Tooltip, toastManager } from "@raystack/apsara";
import styles from "./side-panel.module.css";
import { convertBillingAddressToString } from "../../../../utils/helper";
import Skeleton from "react-loading-skeleton";
import { useContext, useEffect } from "react";
import { CalendarIcon, Pencil1Icon } from "@radix-ui/react-icons";
import { OrganizationContext } from "../contexts/organization-context";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import { CreateCheckoutRequestSchema, FrontierServiceQueries, GetUpcomingInvoiceRequestSchema } from "@raystack/proton/frontier";
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

  const { mutateAsync: createCheckout, isPending: isPortalLoading } =
    useMutation(FrontierServiceQueries.createCheckout, {
      onError: (error) => {
        toastManager.add({
          title: "Something went wrong",
          description: error.rawMessage,
          type: "error",
        });
        console.error("Unable to open billing portal:", error);
      },
    });

  const isBillingPortalDisabled = isPortalLoading || !billingAccountId;

  const handleOpenBillingPortal = async () => {
    if (isBillingPortalDisabled) return;
    const returnUrl = window.location.href;
    try {
      const resp = await createCheckout(
        create(CreateCheckoutRequestSchema, {
          orgId: organizationId,
          cancelUrl: returnUrl,
          successUrl: returnUrl,
          setupBody: { paymentMethod: false, customerPortal: true },
        })
      );
      const checkoutUrl = resp?.checkoutSession?.checkoutUrl;
      if (checkoutUrl) {
        window.location.href = checkoutUrl;
      }
    } catch {
      // the error toast is surfaced by the mutation's onError handler
    }
  };

  return (
    <List className={styles["billing-section"]}>
      <List.Header className={styles["billing-header"]}>
        Billing
        <Tooltip>
          {/* aria-disabled (not the native attribute) keeps the trigger
              hoverable so the tooltip shows even when the action is disabled */}
          <Tooltip.Trigger
            render={
              <IconButton
                size={2}
                aria-label="Update billing details"
                data-test-id="admin-billing-portal-button"
                aria-disabled={isBillingPortalDisabled}
                className={
                  isBillingPortalDisabled
                    ? `${styles["billing-edit"]} ${styles["billing-edit-disabled"]}`
                    : styles["billing-edit"]
                }
                onClick={handleOpenBillingPortal}
              >
                <Pencil1Icon />
              </IconButton>
            }
          />
          <Tooltip.Content>
            {billingAccountId
              ? "Update billing details"
              : "No billing account for this organization"}
          </Tooltip.Content>
        </Tooltip>
      </List.Header>
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
                ? convertBillingAddressToString(billingAccount.address)
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
