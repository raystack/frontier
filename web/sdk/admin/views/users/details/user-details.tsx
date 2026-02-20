import { Flex, EmptyState, Spinner } from "@raystack/apsara";
import { PageTitle } from "../../../components/PageTitle";
import UserIcon from "../../../assets/icons/UsersIcon";
import { UserDetailsLayout } from "./layout";
import { UserProvider } from "./user-context";
import { useQuery } from "@connectrpc/connect-query";
import type { User } from "@raystack/proton/frontier";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { UserDetailsSecurityContent } from "./security/security";

interface UserDetailContentProps {
  user: User;
  refetch: () => void;
}

export const UserDetailContent = ({ user, refetch }: UserDetailContentProps) => {
  return (
    <UserProvider value={{ user, reset: refetch }}>
      <UserDetailsLayout>
        <UserDetailsSecurityContent />
      </UserDetailsLayout>
    </UserProvider>
  );
};

interface UserDetailsByUserIdProps {
  userId: string;
}

export const UserDetailsByUserId = ({ userId }: UserDetailsByUserIdProps) => {
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
    return (
      <Flex justify="center" align="center" style={{ height: "100vh", width: "100%" }}>
        <Spinner size={6} />
      </Flex>
    );
  }

  if (!user?.id) {
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
  }

  return <UserDetailContent user={user} refetch={refetch} />;
};
