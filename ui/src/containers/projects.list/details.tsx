import { Flex, Grid, Text } from "@odpf/apsara";
import { useProject } from ".";

export default function ProjectDetails() {
  const { project } = useProject();
  console.log(project);

  return (
    <Flex
      direction="column"
      css={{
        width: "320px",
        height: "100%",
        padding: "$2",
      }}
    >
      <Text css={{ fontSize: "14px" }}>{project?.name}</Text>
      <Flex direction="column">
        <Grid columns="2" css={{ width: "100%", paddingTop: "$4" }}>
          <Text size={1} css={{ color: "$gray11" }}>
            slug
          </Text>
          <Text size={1}>{project?.slug}</Text>
        </Grid>
        <Grid columns="2" css={{ width: "100%", paddingTop: "$4" }}>
          <Text size={1} css={{ color: "$gray11" }}>
            Created At
          </Text>
          <Text size={1}>
            {new Date(project?.createdAt as Date).toLocaleString("en", {
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
