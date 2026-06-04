import {
  Avatar,
  Button,
  Menu,
  Flex,
  getAvatarColor,
  Text,
} from "@raystack/apsara";
import type React from "react";
import Skeleton from "react-loading-skeleton";
import styles from "./members.module.css";
import { useAddProjectMembers } from "../use-add-project-members";

interface AddMembersDropdownProps {
  projectId: string;
  refetchMembers: () => Promise<void>;
}

function Loader() {
  return (
    <>
      {[...new Array(5)].map((_, i) => (
        <Menu.Item key={i}>
          <Skeleton containerClassName={styles.flex1} />
        </Menu.Item>
      ))}
    </>
  );
}

export function AddMembersDropdown({
  projectId,
  refetchMembers,
}: AddMembersDropdownProps) {
  const { eligibleMembers, isLoading, addMember, setSearchQuery } =
    useAddProjectMembers({
      projectId: projectId,
    });

  function onAddMember(userId: string) {
    return async (e: React.MouseEvent<HTMLDivElement>) => {
      e.stopPropagation();
      await addMember(userId);
      refetchMembers();
    };
  }

  return (
    <Menu
      autocomplete
      onInputValueChange={setSearchQuery}
      autocompleteMode="manual"
    >
      <Menu.Trigger
        render={<Button data-test-id="add-project-member-btn">Add member</Button>}
      />
      <Menu.Content
        side="bottom"
        align="end"
        className={styles["add-member-dropdown"]}
        //  @ts-ignore
        portal={false}
      >
        {isLoading ? (
          <Loader />
        ) : (
          <>
            {eligibleMembers?.slice(0, 7).map((user) => (
              <Menu.Item
                key={user.id}
                onClick={onAddMember(user?.id || "")}
                data-test-id={`admin-add-member-${user.id}`}
                leadingIcon={
                  <Avatar
                    src={user.avatar}
                    fallback={user?.title?.[0] || user?.email?.[0]}
                    radius="full"
                    color={getAvatarColor(user.id || "")}
                  />
                }
                className={styles["add-member-dropdown-member"]}
              >
                <Text className={styles["add-member-dropdown-member-name"]}>
                  {user.title || user.email}
                </Text>
              </Menu.Item>
            ))}
          </>
        )}
      </Menu.Content>
    </Menu>
  );
}
