import { Flex, List, Text, Avatar, Skeleton } from "@raystack/apsara";
import { useMemo } from "react";
import dayjs from "dayjs";
import { type Invitation } from "@raystack/proton/frontier";
import styles from "./side-panel.module.css";
import {
  timestampToDayjs,
  type TimeStamp,
} from "~/admin/utils/connect-timestamp";
import { useOrganizationLookup } from "~/admin/hooks/useOrganizationLookup";
import { useOrganizationRoles } from "~/admin/hooks/useOrganizationRoles";

interface SidePanelInvitationProps {
  data?: Invitation;
  showTitle?: boolean;
  isLoading?: boolean;
}

/*
  Relative expiry text mirrored around now:
  - live → "5 days left", lapsed → "5 days ago"
  - diffs forwards either way, swapping only the suffix, so both sides stay in phase
  - lapsed invites show up at all because the API never filters expires_at
*/
function formatExpiry(expiresAt?: TimeStamp): {
  text: string;
  isExpired: boolean;
} {
  const expires = timestampToDayjs(expiresAt);
  if (!expires) return { text: "-", isExpired: false };

  const now = dayjs();
  const isExpired = !expires.isAfter(now);
  const [from, to] = isExpired ? [expires, now] : [now, expires];
  const suffix = isExpired ? "ago" : "left";

  const days = to.diff(from, "day");
  if (days >= 1) {
    return { text: `${days} day${days === 1 ? "" : "s"} ${suffix}`, isExpired };
  }

  const hours = to.diff(from, "hour");
  if (hours >= 1) {
    return {
      text: `${hours} hour${hours === 1 ? "" : "s"} ${suffix}`,
      isExpired,
    };
  }

  return { text: `Less than an hour ${suffix}`, isExpired };
}

export const SidePanelInvitation = ({
  data,
  showTitle = false,
  isLoading = false,
}: SidePanelInvitationProps) => {
  // Invitation carries only org_id; react-query dedupes repeat lookups.
  const { data: org } = useOrganizationLookup(data?.orgId);

  const { titleById } = useOrganizationRoles(data?.orgId);

  const roleTitles = useMemo(
    () =>
      (data?.roleIds || [])
        .map((roleId) => titleById.get(roleId))
        .filter(Boolean)
        .join(", "),
    [titleById, data?.roleIds],
  );

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

  const orgName = org?.title ?? org?.name ?? data.orgId;
  const { text: expiryText, isExpired } = formatExpiry(data.expiresAt);

  return (
    <List>
      {showTitle && <List.Header>Invitations</List.Header>}
      <List.Item>
        <List.Label className={styles.listLabel}>Name</List.Label>
        <List.Value>
          <Flex gap={3} align="center">
            <Avatar
              src={org?.avatar}
              fallback={orgName?.[0]?.toUpperCase()}
              size={1}
              radius="full"
            />
            <Text className={styles["text-overflow"]}>{orgName}</Text>
          </Flex>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles.listLabel}>Role</List.Label>
        <List.Value>
          <Text className={styles["text-overflow"]}>{roleTitles || "-"}</Text>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles.listLabel}>Invite</List.Label>
        <List.Value>
          <Text variant={isExpired ? "danger" : undefined}>
            {isExpired ? "Expired" : "Pending"}
          </Text>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles.listLabel}>Expiry</List.Label>
        <List.Value>
          <Text>{expiryText}</Text>
        </List.Value>
      </List.Item>
    </List>
  );
};
