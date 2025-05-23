import { Flex, List, Text, Avatar } from "@raystack/apsara/v1";
import dayjs from "dayjs";
import { CalendarIcon } from "@radix-ui/react-icons";
import Skeleton from "react-loading-skeleton";
import { SearchUserOrganizationsResponseUserOrganization } from "~/api/frontier";
import styles from "./side-panel.module.css";
import { MembershipDropdown } from "./membership-dropdown";

interface SidePanelMembershipProps {
  data?: SearchUserOrganizationsResponseUserOrganization;
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
  const orgName = data?.org_title ?? data?.org_name ?? "";

  if (isLoading) {
    return (
      <List.Root>
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
      </List.Root>
    );
  }

  if (!data) return null;

  return (
    <List.Root>
      {showTitle && <List.Header>Membership</List.Header>}
      <List.Item>
        <List.Label minWidth="120px">Name</List.Label>
        <List.Value>
          <Flex gap={3} align="center">
            <Avatar
              src={data?.org_avatar}
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
              {data?.org_joined_on
                ? dayjs(data.org_joined_on).format("DD MMM YYYY")
                : "-"}
            </Text>
          </Flex>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label minWidth="120px">Projects</List.Label>
        <List.Value>{data?.project_count ?? "-"}</List.Value>
      </List.Item>
    </List.Root>
  );
};
