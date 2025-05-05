import { Cross1Icon } from "@radix-ui/react-icons";
import { IconButton, Sheet, SidePanel } from "@raystack/apsara/v1";
import styles from "./edit.module.css";

interface EditBillingPanelProps {
  onClose: () => void;
}

export function EditBillingPanel({ onClose }: EditBillingPanelProps) {
  return (
    <Sheet open>
      <Sheet.Content className={styles["drawer-content"]}>
        <SidePanel
          data-test-id="edit-kyc-panel"
          className={styles["side-panel"]}
        >
          <SidePanel.Header
            title="Edit billing"
            actions={[
              <IconButton
                key="close-billing-panel-icon"
                data-test-id="close-billing-panel-icon"
                onClick={onClose}
              >
                <Cross1Icon />
              </IconButton>,
            ]}
          />
        </SidePanel>
      </Sheet.Content>
    </Sheet>
  );
}
