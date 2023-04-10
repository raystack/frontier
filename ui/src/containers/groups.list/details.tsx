import { Flex, Grid, Text } from "@odpf/apsara";
import { useGroup } from ".";

export default function GroupDetails() {
  const { group } = useGroup();

  return (
    <Flex
      direction="column"
      css={{
        width: "320px",
        height: "100%",
        padding: "$2",
      }}
    >
      <Text css={{ fontSize: "14px" }}>{group?.name}</Text>
      <Flex direction="column">
        <Grid columns="2" css={{ width: "100%", paddingTop: "$4" }}>
          <Text size={1} css={{ color: "$gray11" }}>
            slug
          </Text>
          <Text size={1}>{group?.slug}</Text>
        </Grid>
        <Grid columns="2" css={{ width: "100%", paddingTop: "$4" }}>
          <Text size={1} css={{ color: "$gray11" }}>
            Created At
          </Text>
          <Text size={1}>
            {new Date(group?.createdAt as Date).toLocaleString("en", {
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
