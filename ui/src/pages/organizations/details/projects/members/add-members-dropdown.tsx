import {
  Avatar,
  Button,
  DropdownMenu,
  Flex,
  getAvatarColor,
  Text,
} from "@raystack/apsara/v1";
import { useCallback, useContext, useEffect, useMemo, useState } from "react";
import Skeleton from "react-loading-skeleton";
import styles from "./members.module.css";
import { api } from "~/api";
import { OrganizationContext } from "../../contexts/organization-context";
import { V1Beta1User } from "@raystack/frontier";

interface AddMembersDropdownProps {
  projectId: string;
}

function Loader() {
  return (
    <>
      {[...new Array(5)].map((_, i) => (
        <DropdownMenu.Item key={i}>
          <Skeleton containerClassName={styles.flex1} />
        </DropdownMenu.Item>
      ))}
    </>
  );
}

export function AddMembersDropdown({ projectId }: AddMembersDropdownProps) {
  const { orgMembersMap } = useContext(OrganizationContext);

  const [projectMembersSet, setProjectMembersSet] = useState<Set<string>>(
    new Set(),
  );
  const [isProjectMembersLoading, setIsProjectMembersLoading] = useState(true);

  const fetchMembers = useCallback(async () => {
    try {
      setIsProjectMembersLoading(true);
      const response = await api?.frontierServiceListProjectUsers(projectId);
      const members = response?.data?.users || [];
      const memberSet = new Set(members.map((member) => member.id || ""));
      setProjectMembersSet(memberSet);
    } catch (error) {
      console.error(error);
    } finally {
      setIsProjectMembersLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    fetchMembers();
  }, [fetchMembers]);

  const eligibleUsers = useMemo(() => {
    return Object.values(orgMembersMap).reduce((acc, member) => {
      if (member.id && !projectMembersSet.has(member.id)) {
        acc.push(member);
      }
      return acc;
    }, [] as V1Beta1User[]);
  }, [orgMembersMap, projectMembersSet]);

  const topUsers = useMemo(() => {
    return eligibleUsers.slice(0, 7);
  }, [eligibleUsers]);

  return (
    <DropdownMenu open autocomplete>
      <DropdownMenu.Trigger asChild>
        <Button data-test-id="add-project-member-btn">Add member</Button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        {isProjectMembersLoading ? (
          <Loader />
        ) : (
          topUsers.map((user) => {
            const nameInitial = user.title?.[0] || user?.email?.[0];
            const avatarColor = getAvatarColor(user?.id || "");
            return (
              <DropdownMenu.Item key={user.id}>
                <Flex gap={4} align="center">
                  <Avatar
                    src={user.avatar}
                    fallback={nameInitial}
                    color={avatarColor}
                  />
                  <Text>{user.title || "-"}</Text>
                </Flex>
              </DropdownMenu.Item>
            );
          })
        )}
      </DropdownMenu.Content>
    </DropdownMenu>
  );
}
