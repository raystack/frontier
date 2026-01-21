import { Flex, Text, Grid } from "@raystack/apsara";
import { useRole } from ".";

export default function RoleDetails() {
  const { role } = useRole();

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
