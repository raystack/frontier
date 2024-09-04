import { Flex, Grid, Text } from "@raystack/apsara";
import { NavLink, useParams } from "react-router-dom";
import { useBillingAccount } from ".";
import { BillingAccountAddress } from "@raystack/frontier";
import Skeleton from "react-loading-skeleton";
import { useTokens } from "./tokens/useTokens";

export const converBillingAddressToString = (
  address?: BillingAccountAddress
) => {
  if (!address) return "";
  const { line1, line2, city, state, country, postal_code } = address;
  return [line1, line2, city, state, country, postal_code]
    .filter((v) => v)
    .join(", ");
};

export default function BillingAccountDetails() {
  const { billingaccount } = useBillingAccount();
  let { organisationId, billingaccountId } = useParams();

  const { tokenBalance, isTokensLoading } = useTokens({
    organisationId,
    billingaccountId,
  });

  return (
    <Flex
      direction="column"
      gap="large"
      style={{
        width: "320px",
        height: "calc(100vh - 60px)",
        borderLeft: "1px solid var(--border-base)",
        padding: "var(--pd-16)",
      }}
    >
      <Text size={4}>{billingaccount?.name}</Text>
      <Flex direction="column" gap="large">
        <Grid columns={2} gap="small">
          <Text size={1}>Name</Text>
          <Text size={1}>{billingaccount?.name}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Subscriptions</Text>
          <Text size={1}>
            <NavLink
              to={`/organisations/${organisationId}/billingaccounts/${billingaccountId}/subscriptions`}
              style={{
                display: "flex",
                alignItems: "center",
                flexDirection: "row",
              }}
            >
              Go to subscriptions
            </NavLink>
          </Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Invoices</Text>
          <Text size={1}>
            <NavLink
              to={`/organisations/${organisationId}/billingaccounts/${billingaccountId}/invoices`}
              style={{
                display: "flex",
                alignItems: "center",
                flexDirection: "row",
              }}
            >
              Go to invoices
            </NavLink>
          </Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Tokens</Text>
          {isTokensLoading ? (
            <Skeleton />
          ) : (
            <Text size={1}>
              <NavLink
                to={`/organisations/${organisationId}/billingaccounts/${billingaccountId}/tokens`}
                style={{
                  display: "flex",
                  alignItems: "center",
                  flexDirection: "row",
                }}
              >
                {tokenBalance}
              </NavLink>
            </Text>
          )}
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Organization Id</Text>
          <Text size={1}>{billingaccount?.org_id}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Provider Id</Text>
          <Text size={1}>{billingaccount?.provider_id}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Email</Text>
          <Text size={1}>{billingaccount?.email || "-"}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Address</Text>
          <Text size={1}>
            {converBillingAddressToString(billingaccount?.address) || "-"}
          </Text>
        </Grid>

        <Grid columns={2} gap="small">
          <Text size={1}>Created At</Text>
          <Text size={1}>
            {new Date(billingaccount?.created_at as any).toLocaleString("en", {
              month: "long",
              day: "numeric",
              year: "numeric",
            })}
          </Text>
        </Grid>
      </Flex>
    </Flex>
  );
}

