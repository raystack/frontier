import { Flex, Grid, Text } from "@odpf/apsara";
import { useUser } from ".";

export default function UserDetails() {
  const { user } = useUser();

  return (
    <Flex
      direction="column"
      css={{
        width: "320px",
        height: "100%",
        padding: "$4",
      }}
    >
      <Text css={{ fontSize: "14px" }}>{user?.name}</Text>
      <Flex direction="column">
        <Grid columns="2" css={{ width: "100%", paddingTop: "$4" }}>
          <Text size={1} css={{ color: "$gray11" }}>
            Email
          </Text>
          <Text size={1}>{user?.email}</Text>
        </Grid>
        <Grid columns="2" css={{ width: "100%", paddingTop: "$4" }}>
          <Text
            size={1}
            css={{
              color: "$gray11",
              ...css.row,
            }}
          >
            Created At
          </Text>
          <Text size={1} css={css.row}>
            {new Date(user?.createdAt as Date).toLocaleString("en", {
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

const css = {
  row: { height: "32px", display: "flex", alignItems: "center" },
};
