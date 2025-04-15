import { Cross1Icon } from "@radix-ui/react-icons";
import {
  Button,
  Flex,
  IconButton,
  InputField,
  Label,
  Sheet,
  SidePanel,
  Switch,
  Text,
} from "@raystack/apsara/v1";
import styles from "./kyc.module.css";

export function EditKYCPanel() {
  return (
    <Sheet open>
      <Sheet.Content className={styles["drawer-content"]}>
        <SidePanel
          data-test-id="edit-kyc-panel"
          className={styles["side-panel"]}
        >
          <SidePanel.Header
            title={"Edit KYC"}
            actions={[
              <IconButton
                key={"close-kyc-panel-icon"}
                data-test-id="close-kyc-panel-icon"
              >
                <Cross1Icon />
              </IconButton>,
            ]}
          />
          <Flex direction={"column"} justify={"between"}>
            <Flex
              direction={"column"}
              gap={5}
              className={styles["side-panel-content"]}
            >
              <Text size="small" weight={"medium"}>
                KYC Details
              </Text>
              <Flex justify={"between"}>
                <Label>Mark KYC as verified</Label>
                <Switch />
              </Flex>
              <Flex>
                <InputField label="Document Link" />
              </Flex>
            </Flex>
            <Flex className={styles["side-panel-footer"]} gap={3}>
              <Button variant={"outline"} color="neutral">
                Cancel
              </Button>
              <Button>Save</Button>
            </Flex>
          </Flex>
        </SidePanel>
      </Sheet.Content>
    </Sheet>
  );
}
