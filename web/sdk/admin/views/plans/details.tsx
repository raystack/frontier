import { Flex, Text, Grid } from "@raystack/apsara";
import type { Plan } from "@raystack/proton/frontier";
import { timestampToDayjs } from "../../utils/connect-timestamp";

export default function PlanDetails({ plan }: { plan: Plan | null }) {

  return (
    <Flex direction="column" gap={9}>
      <Text size="regular">{plan?.name}</Text>
      <Flex direction="column" gap={9}>
        <Grid columns={2} gap={3}>
          <Text size="mini">Name</Text>
          <Text size="mini">{plan?.name}</Text>
        </Grid>
        <Grid columns={2} gap={3}>
          <Text size="mini">Interval</Text>
          <Text size="mini">{plan?.interval}</Text>
        </Grid>
        <Grid columns={2} gap={3}>
          <Text size="mini">Created At</Text>
          <Text size="mini">
            {(() => {
              const date = timestampToDayjs(plan?.createdAt);
              return date ? date.format("DD MMM YYYY") : "-";
            })()}
          </Text>
        </Grid>
      </Flex>
    </Flex>
  );
}
