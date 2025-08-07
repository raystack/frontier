import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { Button, EmptyState, Flex } from "@raystack/apsara";
import { useMutation } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";

export default function UnauthorizedState() {
  const logoutMutation = useMutation(FrontierServiceQueries.authLogout, {
    onSuccess: () => {
      window.location.href = "/";
      window.location.reload();
    },
  });
  return (
    <Flex style={{ height: "100vh" }}>
      <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="Unauthorized"
        subHeading="You dont have access to view this page"
        primaryAction={
          <Button
            onClick={() => logoutMutation.mutate({})}
            data-test-id="admin-ui-unauthorized-screen-logout-btn"
          >
            Logout
          </Button>
        }
      ></EmptyState>
    </Flex>
  );
}
