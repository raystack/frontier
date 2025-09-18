import { Flex, List, Text, Avatar } from "@raystack/apsara";
import dayjs from "dayjs";
import { CalendarIcon } from "@radix-ui/react-icons";
import Skeleton from "react-loading-skeleton";
import { type SearchUserOrganizationsResponse_UserOrganization } from "@raystack/proton/frontier";
import styles from "./side-panel.module.css";
import { MembershipDropdown } from "./membership-dropdown";
import { timestampToDate, isNullTimestamp } from "~/utils/connect-timestamp";

interface SidePanelMembershipProps {
  data?: SearchUserOrganizationsResponse_UserOrganization;
  showTitle?: boolean;
  isLoading?: boolean;
  onReset?: () => void;
}

export const SidePanelMembership = ({
  data,
  showTitle = false,
  isLoading = false,
  onReset,
}: SidePanelMembershipProps) => {
  const orgName = data?.orgTitle ?? data?.orgName ?? "";
  if (isLoading) {
    return (
      <List>
        <Flex className={styles["loader-header"]}>
          <Skeleton />
        </Flex>
        {[...Array(4)].map((_, index) => (
          <List.Item key={index}>
            <List.Value>
              <Skeleton height="100%" />
            </List.Value>
          </List.Item>
        ))}
      </List>
    );
  }

  if (!data) return null;

  return (
    <List>
      {showTitle && <List.Header>Membership</List.Header>}
      <List.Item>
        <List.Label minWidth="120px">Name</List.Label>
        <List.Value>
          <Flex gap={3} align="center">
            <Avatar
              src={data?.orgAvatar}
              fallback={orgName?.[0]?.toUpperCase()}
              size={1}
              radius="full"
            />
            <Text>{orgName}</Text>
          </Flex>
        </List.Value>
      </List.Item>
      <List.Item className={styles["dropdown-item"]}>
        <List.Label minWidth="112px">Role</List.Label>
        <List.Value>
          <MembershipDropdown data={data} onReset={onReset} />
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="120px">Joined on</List.Label>
        <List.Value>
          <Flex gap={3}>
            <CalendarIcon />
            <Text>
              {data?.orgJoinedOn && !isNullTimestamp(data.orgJoinedOn)
                ? dayjs(timestampToDate(data.orgJoinedOn)).format("DD MMM YYYY")
                : "-"}
            </Text>
          </Flex>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="120px">Projects</List.Label>
        <List.Value>
          {data?.projectCount ? Number(data.projectCount) : "-"}
        </List.Value>
      </List.Item>
    </List>
  );
};
