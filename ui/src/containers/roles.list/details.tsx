import { Flex, Grid, Text } from "@raystack/apsara";
import { useRole } from ".";

export default function RoleDetails() {
  const { role } = useRole();

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
      <Text size={4}>{role?.name}</Text>
      <Flex direction="column" gap="large">
        <Grid columns={2} gap="small">
          <Text size={1}>Name</Text>
          <Text size={1}>{role?.name}</Text>
        </Grid>
      </Flex>
    </Flex>
  );
}