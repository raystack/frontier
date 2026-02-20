import { Flex, Text, Grid } from "@raystack/apsara";
import type { Plan } from "@raystack/proton/frontier";
import { timestampToDate } from "../../utils/connect-timestamp";

export default function PlanDetails({ plan }: { plan: Plan | null }) {

  return (
    <Flex direction="column" gap={9}>
      <Text size={4}>{plan?.name}</Text>
      <Flex direction="column" gap={9}>
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
            {(() => {
              const date = timestampToDate(plan?.createdAt);
              return date ? date.toLocaleString("en", {
                month: "long",
                day: "numeric",
                year: "numeric",
              }) : "-";
            })()}
          </Text>
        </Grid>
      </Flex>
    </Flex>
  );
}
