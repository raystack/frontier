import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { Button, EmptyState, Flex } from "@raystack/apsara/v1";
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
      <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="Unauthorized"
        subHeading="You dont have access to view this page"
        primaryAction={
          <Button
            onClick={logout}
            data-test-id="admin-ui-unauthorized-screen-logout-btn"
          >
            Logout
          </Button>
        }
      ></EmptyState>
    </Flex>
  );
}
