import { Grid } from "@raystack/apsara";
import { Flex, Text } from "@raystack/apsara/v1";
import { useGroup } from ".";

export default function GroupDetails() {
  const { group } = useGroup();

  return (
    <Flex
      direction="column"
      gap="large"
      style={{
        width: "320px",
        height: "calc(100vh - 60px)",
        borderLeft: "1px solid var(--rs-color-border-base-primary)",
        padding: "var(--rs-space-5)",
      }}
    >
      <Text size={4}>{group?.name}</Text>
      <Flex direction="column" gap="large">
        <Grid columns={2} gap="small">
          <Text size={1}>Name</Text>
          <Text size={1}>{group?.name}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Organization Id</Text>
          <Text size={1}>{group?.org_id}</Text>
        </Grid>
        <Grid columns={2} gap="small">
          <Text size={1}>Created At</Text>
          <Text size={1}>
            {new Date(group?.created_at as any).toLocaleString("en", {
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
