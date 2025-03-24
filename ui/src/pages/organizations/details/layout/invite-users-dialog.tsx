import { Button, Dialog } from "@raystack/apsara/v1";
import React from "react";

export const InviteUsersDialog = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  return (
    <Dialog>
      <Dialog.Trigger asChild>{children}</Dialog.Trigger>
      <Dialog.Content width={600}>
        <Dialog.Header>
          <Dialog.Title>Invite user</Dialog.Title>
          <Dialog.CloseButton data-test-id="invite-users-close-button" />
        </Dialog.Header>
        <Dialog.Body>
          <span />
        </Dialog.Body>
        <Dialog.Footer>
          <Button data-test-id="invite-users-invite-button">Invite</Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
