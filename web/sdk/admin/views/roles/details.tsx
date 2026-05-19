import { Flex, Text, Grid } from "@raystack/apsara-v1";
import type { Role } from "@raystack/proton/frontier";

export default function RoleDetails({ role }: { role: Role | null }) {
  return (
    <Flex direction="column" gap={9}>
      <Text size="regular">{role?.name}</Text>
      <Flex direction="column" gap={9}>
        <Grid columns={2} gap={3}>
          <Text size="mini">Name</Text>
          <Text size="mini">{role?.name}</Text>
        </Grid>
      </Flex>
    </Flex>
  );
}
