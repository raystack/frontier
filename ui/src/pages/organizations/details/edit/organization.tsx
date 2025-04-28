import { useContext } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import styles from "./edit.module.css";
import {
  Button,
  Flex,
  IconButton,
  InputField,
  Sheet,
  SidePanel,
  Text,
} from "@raystack/apsara/v1";
import { Cross1Icon } from "@radix-ui/react-icons";
import { AvatarUpload } from "@raystack/frontier/react";

export function EditOrganizationPanel({ onClose }: { onClose: () => void }) {
  const { organization } = useContext(OrganizationContext);

  return (
    <Sheet open>
      <Sheet.Content className={styles["drawer-content"]}>
        <SidePanel
          data-test-id="edit-org-panel"
          className={styles["side-panel"]}
        >
          <SidePanel.Header
            title="Edit organization"
            actions={[
              <IconButton
                key="close-edit-org-panel-icon"
                data-test-id="close-edit-org-panel-icon"
                onClick={onClose}
              >
                <Cross1Icon />
              </IconButton>,
            ]}
          />
          <form className={styles["side-panel-form"]}>
            <Flex
              direction="column"
              gap={5}
              className={styles["side-panel-content"]}
            >
              <Flex align={"center"} gap={"medium"} style={{ width: "100%" }}>
                <AvatarUpload />
                <Text>Pick a logo for your organization</Text>
              </Flex>
              <InputField label="Organization title" />
              <InputField
                prefix="aurora.pixxel.space"
                label="Organization URL"
                helperText="This will be your organization unique web address"
              />
              <InputField type="number" label="Organization size" />
              <InputField label="Organization industry" />
              <InputField label="Country" />
            </Flex>
            <Flex className={styles["side-panel-footer"]} gap={3}>
              <Button
                variant="outline"
                color="neutral"
                onClick={onClose}
                data-test-id="cancel-edit-org-button"
              >
                Cancel
              </Button>
              <Button
                data-test-id="save-edit-org-button"
                type="submit"
                loaderText="Saving..."
              >
                Save
              </Button>
            </Flex>
          </form>
        </SidePanel>
      </Sheet.Content>
    </Sheet>
  );
}
