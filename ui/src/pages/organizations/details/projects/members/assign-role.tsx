import {
  Button,
  Checkbox,
  Dialog,
  Flex,
  Label,
  Text,
} from "@raystack/apsara/v1";
import styles from "./members.module.css";
import { useState } from "react";

export const AssignRole = () => {
  const [checked, setChecked] = useState(false);
  function onCheckedChange(e: any) {
    setChecked((prev) => !prev);
  }
  return (
    <Dialog open>
      <Dialog.Content
        width={400}
        overlayClassName={styles["action-dialog-overlay"]}
        className={styles["action-dialog-content"]}
      >
        <Dialog.Header>
          <Dialog.Title>Assign Role</Dialog.Title>
          <Dialog.CloseButton data-test-id="assign-role-close-button" />
        </Dialog.Header>
        <Dialog.Body>
          <Flex direction="column" gap={7}>
            <Text variant="secondary">
              Taking this action may result in changes in the role which might
              lead to changes in access of the user.
            </Text>
            <Flex direction="column" gap={4}>
              <Flex gap={3}>
                <Checkbox
                  id="abcd"
                  checked={checked}
                  onCheckedChange={onCheckedChange}
                />
                <Label htmlFor="abcd">Owner</Label>
              </Flex>
            </Flex>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer>
          <Dialog.Close asChild>
            <Button
              variant="outline"
              color="neutral"
              data-test-id="assign-role-cancel-button"
            >
              Cancel
            </Button>
          </Dialog.Close>
          <Button data-test-id="assign-role-update-button">Update</Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
