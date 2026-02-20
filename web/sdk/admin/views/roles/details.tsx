import { Flex, Text, Grid } from "@raystack/apsara";
import type { Role } from "@raystack/proton/frontier";

export default function RoleDetails({ role }: { role: Role | null }) {
  return (
    <Flex direction="column" gap={9}>
      <Text size={4}>{role?.name}</Text>
      <Flex direction="column" gap={9}>
        <Grid columns={2} gap="small">
          <Text size={1}>Name</Text>
          <Text size={1}>{role?.name}</Text>
        </Grid>
      </Flex>
    </Flex>
  );
}
