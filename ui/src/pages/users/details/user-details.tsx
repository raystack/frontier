import { Outlet, useParams } from "react-router-dom";
import { Flex, EmptyState } from "@raystack/apsara";
import LoadingState from "~/components/states/Loading";
import PageTitle from "~/components/page-title";
import UserIcon from "~/assets/icons/users.svg?react";
import { UserDetailsLayout } from "./layout";
import { UserProvider } from "./user-context";
import { useQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";

export const UserDetails = () => {
  const { userId } = useParams();

  const { data, isLoading, refetch } = useQuery(
    AdminServiceQueries.searchUsers,
    { query: { search: userId } },
    {
      enabled: !!userId,
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );
  const user = data?.users?.[0];

  if (isLoading) {
    return <LoadingState />;
  }

  if (!user?.id)
    return (
      <Flex
        style={{ height: "100vh", width: "100%" }}
        align="center"
        justify="center">
        <PageTitle title="User not found" />
        <EmptyState
          icon={<UserIcon />}
          heading="User not found"
          subHeading="The user you are looking for does not exist."
        />
      </Flex>
    );

  return (
    <UserProvider value={{ user, reset: refetch }}>
      <UserDetailsLayout>
        <Outlet />
      </UserDetailsLayout>
    </UserProvider>
  );
};
