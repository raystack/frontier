import { Grid } from "@raystack/apsara";
import { Flex, Text } from "@raystack/apsara/v1";
import { usePlan } from ".";
export default function PlanDetails() {
  const { plan } = usePlan();

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
      <Text size={4}>{plan?.name}</Text>
      <Flex direction="column" gap="large">
        <Grid columns={2} gap="small">
          <Text size={1}>Name</Text>
          <Text size={1}>{plan?.name}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Interval</Text>
          <Text size={1}>{plan?.interval}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Created At</Text>
          <Text size={1}>
            {new Date(plan?.created_at as any).toLocaleString("en", {
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
