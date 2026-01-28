import { Flex, IconButton, SidePanel } from "@raystack/apsara";
import {
  Cross2Icon,
  DesktopIcon,
  GlobeIcon,
  TransformIcon,
} from "@radix-ui/react-icons";
import { List } from "@raystack/apsara";
import styles from "./list.module.css";
import { AuditRecord } from "@raystack/proton/frontier";
import { ACTOR_TYPES } from "../util";
import { timestampToDate } from "~/utils/connect-timestamp";
import dayjs from "dayjs";
import MapIcon from "~/assets/icons/map.svg?react";
import SidePanelLogDialog from "./sidepanel-log-dialog";
import ActorCell from "./actor-cell";
import SidepanelListItemLink from "./sidepanel-list-link";
import { isZeroUUID } from "~/utils/helper";
import SidepanelListId from "./sidepanel-list-id";

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
  onClose,
  ...rest
}: SidePanelDetailsProps) {
  const { actor, event, resource, occurredAt, id, orgId, orgName, target } =
    rest;
  const date = dayjs(timestampToDate(occurredAt));

  const session = actor?.metadata?.context as AuditSessionContext;
  const location =
    (session && `${session.Location.City}, ${session.Location.Country}`) || "-";

  return (
    <SidePanel
      data-test-id="admin-user-details-sidepanel"
      className={styles["side-panel"]}>
      <SidePanel.Header
        title="Audit log details"
        actions={[
          <SidePanelLogDialog key="show-audit-json-dialog" {...rest} />,
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
          <SidepanelListItemLink
            isLink={actor?.type !== ACTOR_TYPES.SYSTEM}
            href={`/users/${actor?.id}`}
            label="Actor"
            data-test-id="actor-link">
            <ActorCell value={actor!} size="small" maxLength={12} />
          </SidepanelListItemLink>
          <SidepanelListItemLink
            isLink={!!orgId && !isZeroUUID(orgId)}
            href={`/organizations/${orgId}`}
            label="Organization"
            data-test-id="actor-link">
            {orgName || "-"}
          </SidepanelListItemLink>
          <List.Item>
            <List.Label minWidth="120px">Action</List.Label>
            <List.Value>{event}</List.Value>
          </List.Item>
          <List.Item>
            <List.Label minWidth="120px">Resource</List.Label>
            <List.Value>{resource?.name || "-"}</List.Value>
          </List.Item>
          <List.Item>
            <List.Label minWidth="120px">Type</List.Label>
            <List.Value>{resource?.type || "-"}</List.Value>
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
            <SidepanelListId id={id} />
          </List.Item>
          {target && (
            <>
              <List.Item>
                <List.Label minWidth="120px">Target ID</List.Label>
                <SidepanelListId id={target?.id} />
              </List.Item>
              <List.Item>
                <List.Label minWidth="120px">Target Type</List.Label>
                <List.Value>{target?.type || "-"}</List.Value>
              </List.Item>
            </>
          )}
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
