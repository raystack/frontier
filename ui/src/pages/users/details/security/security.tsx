import { Flex } from "@raystack/apsara/v1";
import PageTitle from "~/components/page-title";
import { useUser } from "../user-context";

export const UserDetailsSecurityPage = () => {
  const user = useUser();

  const title = `Security | ${user?.email} | Users`;

  return (
    <Flex justify="center">
      <PageTitle title={title} />
    </Flex>
  );
};
