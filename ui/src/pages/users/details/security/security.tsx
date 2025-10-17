import { Flex, Separator, Text } from "@raystack/apsara";
import PageTitle from "~/components/page-title";
import { useUser } from "../user-context";
import styles from "./security.module.css";
import { BlockUserDialog } from "./block-user";
import { UserSessions } from "./sessions";

export const UserDetailsSecurityPage = () => {
  const { user } = useUser();

  const title = `Security | ${user?.email} | Users`;

  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
      
      <Flex className={styles["content"]} direction="column" gap={9}>
        <UserSessions />
        <Separator />
        <Flex gap={5} justify="between" className={styles["blockUserSection"]}>
          <Flex direction="column" gap={3}>
            <Text size={5}>Block user</Text>
            <Text size={3} variant="secondary">
              Block user access to safeguard platform integrity and prevent
              unauthorized activities.
            </Text>
          </Flex>

          <BlockUserDialog />
        </Flex>
      </Flex>
    </Flex>
  );
};
