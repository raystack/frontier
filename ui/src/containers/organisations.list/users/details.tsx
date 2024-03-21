import { Flex, Text } from "@raystack/apsara";
import { useUser } from ".";

export default function OrgUserDetails() {
  const { user } = useUser();
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
      <Text size={4}>{user?.title || user?.email}</Text>
    </Flex>
  );
}
