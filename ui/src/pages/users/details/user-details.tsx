import { useCallback, useEffect, useState } from "react";
import { Outlet, useParams } from "react-router-dom";
import { Flex, EmptyState } from "@raystack/apsara";
import { V1Beta1User } from "~/api/frontier";
import { api } from "~/api";
import LoadingState from "~/components/states/Loading";
import PageTitle from "~/components/page-title";
import UserIcon from "~/assets/icons/users.svg?react";
import { UserDetailsLayout } from "./layout";
import { UserProvider } from "./user-context";

export const UserDetails = () => {
  const { userId } = useParams();
  const [user, setUser] = useState<V1Beta1User>();
  const [isLoading, setIsLoading] = useState(true);

  const fetchUser = useCallback(async (id: string) => {
    try {
      setIsLoading(true);
      const response = await api?.adminServiceSearchUsers({
        search: id,
      });
      setUser(response.data?.users?.[0]);
    } catch (error) {
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const resetUser = () => {
    if (userId) fetchUser(userId);
  };

  useEffect(resetUser, [userId, fetchUser]);

  if (isLoading) {
    return <LoadingState />;
  }

  if (!user?.id)
    return (
      <Flex
        style={{ height: "100vh", width: "100%" }}
        align="center"
        justify="center"
      >
        <PageTitle title="User not found" />
        <EmptyState
          icon={<UserIcon />}
          heading="User not found"
          subHeading="The user you are looking for does not exist."
        />
      </Flex>
    );

  return (
    <UserProvider value={{ user, reset: resetUser }}>
      <UserDetailsLayout>
        <Outlet />
      </UserDetailsLayout>
    </UserProvider>
  );
};
