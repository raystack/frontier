import { ComponentProps, useState } from "react";
import { Button, Dialog, toast } from "@raystack/apsara";
import { api } from "~/api";
import { useUser } from "../user-context";
import { getUserName } from "../../util";

type ButtonColorType = ComponentProps<typeof Button>["color"];

export const BlockUserDialog = () => {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { user, reset } = useUser();

  const isActive = user?.state === "enabled";
  const userName = getUserName(user);

  async function blockUser() {
    if (!user?.id) return;
    try {
      setIsSubmitting(true);
      await api?.frontierServiceDisableUser(user.id, {});
      toast.success("User blocked successfully");
      reset?.();
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to block user", error);
    } finally {
      setIsSubmitting(false);
    }
  }

  async function unblockUser() {
    if (!user?.id) return;
    try {
      setIsSubmitting(true);
      await api?.frontierServiceEnableUser(user.id, {});
      toast.success("User unblocked successfully");
      reset?.();
      onOpenChange(false);
    } catch (error) {
      console.error("Failed to unblock user", error);
    } finally {
      setIsSubmitting(false);
    }
  }

  const onOpenChange = (value: boolean) => {
    setIsDialogOpen(value);
  };

  const config = isActive
    ? {
        onClick: blockUser,
        btnColor: "danger" as ButtonColorType,
        btnText: "Block",
        dialogTitle: "Block user",
        dialogDescription: `Blocking this user will permanently restrict access to its content, disable communication, and prevent any future interactions. Are you sure you want to suspend ${userName}?`,
        dialogConfirmText: "Block",
        dialogConfirmLoadingText: "Blocking...",
      }
    : {
        onClick: unblockUser,
        btnColor: "accent" as ButtonColorType,
        btnText: "Unblock",
        dialogTitle: "Unblock user",
        dialogDescription: `Unblocking this user will restore access to its content, enable communication, and allow future interactions. Are you sure you want to unblock ${userName}?`,
        dialogConfirmText: "Unblock",
        dialogConfirmLoadingText: "Unblocking...",
      };

  return (
    <Dialog open={isDialogOpen} onOpenChange={onOpenChange}>
      <Dialog.Trigger asChild>
        <Button
          color={config.btnColor}
          size="small"
          data-test-id="admin-ui-security-block-user"
        >
          {config.btnText}
        </Button>
      </Dialog.Trigger>
      <Dialog.Content width={400} ariaLabel="Block user">
        <Dialog.Body>
          <Dialog.Title>{config.dialogTitle}</Dialog.Title>
          <Dialog.Description>{config.dialogDescription}</Dialog.Description>
        </Dialog.Body>
        <Dialog.Footer>
          <Dialog.Close asChild>
            <Button
              color="neutral"
              variant="outline"
              data-test-id="admin-ui-security-block-user-cancel"
            >
              Cancel
            </Button>
          </Dialog.Close>
          <Button
            color={config.btnColor}
            data-test-id="admin-ui-security-block-user-submit"
            loading={isSubmitting}
            loaderText={config.dialogConfirmLoadingText}
            onClick={config.onClick}
          >
            {config.dialogConfirmText}
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
