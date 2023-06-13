import { Flex, Grid, Text } from "@raystack/apsara";
import { useRole } from ".";

export default function RoleDetails() {
  const { role } = useRole();

  return (
    <Flex
      direction="column"
      css={{
        width: "320px",
        height: "100%",
        padding: "$4",
      }}
    >
      <Text css={{ fontSize: "14px" }}>{role?.name}</Text>
      <Flex direction="column">
        <Grid columns="2" css={{ width: "100%", paddingTop: "$4" }}>
          <Text size={1} css={{ color: "$gray11" }}>
            Name
          </Text>
          <Text size={1}>{role?.name}</Text>
        </Grid>
        <Grid columns="2" css={{ width: "100%", paddingTop: "$4" }}>
          <Text
            size={1}
            css={{
              color: "$gray11",
              ...css.row,
            }}
          >
            Types
          </Text>
          <Text size={1} css={css.row}>
            {role?.types}
          </Text>
        </Grid>
      </Flex>
    </Flex>
  );
}

const css = {
  row: { height: "32px", display: "flex", alignItems: "center" },
};
