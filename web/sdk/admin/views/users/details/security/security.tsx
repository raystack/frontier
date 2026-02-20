import { Flex, Separator, Text } from "@raystack/apsara";
import { PageTitle } from "../../../../components/PageTitle";
import { useUser } from "../user-context";
import { BlockUserDialog } from "./block-user";
import { UserSessions } from "./sessions";
import styles from "./security.module.css";

export const UserDetailsSecurityContent = () => {
  const { user } = useUser();

  const title = `Security | ${user?.email} | Users`;

  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />

      <Flex className={styles["content"]} direction="column" gap={9}>
        <UserSessions />
        <Separator />
        <Flex gap={5} justify="between">
          <Flex direction="column" gap={3}>
            <Text size="large" weight="medium">Block user</Text>
            <Text size="regular" variant="secondary">
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
