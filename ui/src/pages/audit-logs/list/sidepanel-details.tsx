import { IconButton, SidePanel } from "@raystack/apsara";
import { Cross2Icon } from "@radix-ui/react-icons";
import { List } from "@raystack/apsara";
import styles from "./list.module.css";
import { AuditRecord } from "@raystack/proton/frontier";

type SidePanelDetailsProps = Partial<AuditRecord> & {
  onClose: () => void;
};

export default function SidePanelDetails({
  actor,
  onClose,
}: SidePanelDetailsProps) {
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
            <List.Value>{actor?.name}</List.Value>
          </List.Item>
        </List>
      </SidePanel.Section>
    </SidePanel>
  );
}
