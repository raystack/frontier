import {
  Avatar,
  CopyButton,
  Flex,
  getAvatarColor,
  IconButton,
  SidePanel,
  Text,
  Tooltip,
} from "@raystack/apsara";
import {
  Cross2Icon,
  DesktopIcon,
  GlobeIcon,
  TransformIcon,
} from "@radix-ui/react-icons";
import { List } from "@raystack/apsara";
import styles from "./list.module.css";
import { AuditRecord } from "@raystack/proton/frontier";
import { isAuditLogActorServiceUser } from "../util";
import { getAuditLogActorName } from "../util";
import serviceUserIcon from "~/assets/images/service-user.jpg";
import { timestampToDate } from "~/utils/connect-timestamp";
import dayjs from "dayjs";
import MapIcon from "~/assets/icons/map.svg?react";

type SidePanelDetailsProps = Partial<AuditRecord> & {
  onClose: () => void;
};

type AuditSessionContext = {
  Browser: string;
  IpAddress: string;
  Location: {
    City: string;
    Country: string;
  };
  OperatingSystem: string;
};

export default function SidePanelDetails({
  actor,
  event,
  resource,
  occurredAt,
  onClose,
  id,
}: SidePanelDetailsProps) {
  const name = getAuditLogActorName(actor);
  const isServiceUser = isAuditLogActorServiceUser(actor);
  const date = dayjs(timestampToDate(occurredAt));

  const session = actor?.metadata?.context as AuditSessionContext;
  const location =
    (session && `${session.Location.City}, ${session.Location.Country}`) || "-";

  return (
    <SidePanel
      data-test-id="admin-ui-user-details-sidepanel"
      className={styles["side-panel"]}>
      <SidePanel.Header
        title="Audit log details"
        actions={[
          <IconButton
            size={3}
            key="close-sidepanel-icon"
            data-test-id="close-sidepanel-icon"
            onClick={onClose}>
            <Cross2Icon />
          </IconButton>,
        ]}
      />
      <SidePanel.Section>
        <List>
          <List.Header>Overview</List.Header>
          <List.Item>
            <List.Label minWidth="120px">Actor</List.Label>
            <List.Value>
              <Flex gap={3} align="center">
                <Avatar
                  size={1}
                  fallback={name?.[0]?.toUpperCase()}
                  color={getAvatarColor(actor?.id ?? "")}
                  radius="full"
                  src={isServiceUser ? serviceUserIcon : undefined}
                />
                <Text size="regular">{name}</Text>
              </Flex>
            </List.Value>
          </List.Item>
          <List.Item>
            <List.Label minWidth="120px">Action</List.Label>
            <List.Value className={styles.capitalize}>{event}</List.Value>
          </List.Item>
          <List.Item>
            <List.Label minWidth="120px">Resource</List.Label>
            <List.Value className={styles.capitalize}>
              {resource?.name || "-"}
            </List.Value>
          </List.Item>
          <List.Item>
            <List.Label minWidth="120px">Type</List.Label>
            <List.Value className={styles.capitalize}>
              {resource?.type || "-"}
            </List.Value>
          </List.Item>
          <List.Item>
            <List.Label minWidth="120px">Date</List.Label>
            <List.Value>{date.format("DD MMM YYYY")}</List.Value>
          </List.Item>
          <List.Item>
            <List.Label minWidth="120px">Time</List.Label>
            <List.Value>{date.format("hh:mm A")}</List.Value>
          </List.Item>
          <List.Item>
            <List.Label minWidth="120px">ID</List.Label>
            <List.Value>
              <Flex gap={3} style={{ width: "100%" }}>
                <CopyButton text={id || ""} data-test-id="copy-button" />
                <Tooltip message={id || ""}>
                  <Text className={styles["text-overflow"]} weight="medium">
                    {id}
                  </Text>
                </Tooltip>
              </Flex>
            </List.Value>
          </List.Item>
        </List>
      </SidePanel.Section>
      {session && (
        <SidePanel.Section>
          <List>
            <List.Header>Session</List.Header>
            <List.Item>
              <List.Label minWidth="120px">IP Address</List.Label>
              <List.Value>
                <Flex gap={3} align="center">
                  <GlobeIcon width={16} height={16} />
                  {session.IpAddress.split(",")[0]}
                </Flex>
              </List.Value>
            </List.Item>
            <List.Item>
              <List.Label minWidth="120px">Location</List.Label>
              <List.Value>
                <Flex gap={3} align="center">
                  <MapIcon width={16} height={16} />
                  {location.length > 2 ? location : "-"}
                </Flex>
              </List.Value>
            </List.Item>
            <List.Item>
              <List.Label minWidth="120px">Browser</List.Label>
              <List.Value>
                <Flex gap={3} align="center">
                  <DesktopIcon width={16} height={16} />
                  {session.Browser}
                </Flex>
              </List.Value>
            </List.Item>
            <List.Item>
              <List.Label minWidth="120px">Operating System</List.Label>
              <List.Value>
                <Flex gap={3} align="center">
                  <TransformIcon width={16} height={16} />
                  {session.OperatingSystem}
                </Flex>
              </List.Value>
            </List.Item>
          </List>
        </SidePanel.Section>
      )}
    </SidePanel>
  );
}
