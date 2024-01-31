import { Flex, Grid, Text } from "@raystack/apsara";
import { NavLink, useParams } from "react-router-dom";
import { usebillingaccount } from ".";

export default function BillingAccountDetails() {
  const { billingaccount } = usebillingaccount();
  let { organisationId, billingaccountId } = useParams();
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
          <Text size={1}>Organization Id</Text>
          <Text size={1}>{billingaccount?.org_id}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Provider</Text>
          <Text size={1}>{billingaccount?.provider}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>State</Text>
          <Text size={1}>{billingaccount?.state}</Text>
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

const css = {
  row: { height: "32px", display: "flex", alignItems: "center" },
};
