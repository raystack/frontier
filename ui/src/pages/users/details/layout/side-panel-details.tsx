import { Flex, List, Text, CopyButton, Tooltip } from "@raystack/apsara/v1";
import dayjs from "dayjs";
import { CalendarIcon } from "@radix-ui/react-icons";
import styles from "./side-panel.module.css";
import { UserState, USER_STATES } from "../../util";
import { useUser } from "../user-context";

export const SidePanelDetails = () => {
  const user = useUser();

  return (
    <List.Root>
      <List.Header>User Details</List.Header>
      <List.Item>
        <List.Label minWidth="120px">ID</List.Label>
        <List.Value>
          <Flex gap={3} style={{ width: "100%" }}>
            <CopyButton text={user.id || ""} data-test-id="copy-button" />
            <Tooltip message={user.id}>
              <Text className={styles["text-overflow"]}>{user.id}</Text>
            </Tooltip>
          </Flex>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="120px">Email</List.Label>
        <List.Value>
          <Text>{user.email}</Text>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="120px">Created on</List.Label>
        <List.Value>
          <Flex gap={3}>
            <CalendarIcon />
            <Text>{dayjs(user?.created_at).format("DD MMM YYYY")}</Text>
          </Flex>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="120px">Status</List.Label>
        <List.Value>
          {user?.state ? USER_STATES?.[user.state as UserState] : "-"}
        </List.Value>
      </List.Item>
    </List.Root>
  );
};
