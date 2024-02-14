import { Button, EmptyState, Flex, Text } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";

export default function UnauthorizedState() {
  const { client } = useFrontier();
  async function logout() {
    await client?.frontierServiceAuthLogout();
    window.location.href = "/";
    window.location.reload();
  }
  return (
    <Flex style={{ height: "100vh" }}>
      <EmptyState>
        <Text size={5}>Unauthorized</Text>
        <Text>You dont have access to view this page</Text>
        <Button variant={"primary"} onClick={logout}>
          Logout
        </Button>
      </EmptyState>
    </Flex>
  );
}
