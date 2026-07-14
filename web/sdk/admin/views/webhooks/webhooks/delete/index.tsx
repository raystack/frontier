import { AlertDialog, Button, toastManager } from "@raystack/apsara";
import type { useMutation } from "@connectrpc/connect-query";
import { handleConnectError } from "~/utils/error";

interface DeleteWebhookDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  webhookId: string;
  webhookDescription?: string;
  deleteWebhookMutation: ReturnType<typeof useMutation>;
}

export function DeleteWebhookDialog({
  isOpen,
  onOpenChange,
  webhookId,
  webhookDescription,
  deleteWebhookMutation,
}: DeleteWebhookDialogProps) {
  const handleDelete = async () => {
    try {
      await deleteWebhookMutation.mutateAsync({ id: webhookId });
      toastManager.add({ title: "Webhook deleted", type: "success" });
      onOpenChange(false);
    } catch (err) {
      console.error("Failed to delete webhook:", err);
      handleConnectError(err, {
        PermissionDenied: () =>
          toastManager.add({
            title: "You don't have permission to perform this action",
            type: "error",
          }),
        Default: (e) =>
          toastManager.add({
            title: "Failed to delete webhook",
            description: e.rawMessage,
            type: "error",
          }),
      });
    }
  };

  return (
    <AlertDialog open={isOpen} onOpenChange={onOpenChange}>
      <AlertDialog.Content width={600}>
        <AlertDialog.Header>
          <AlertDialog.Title>Delete Webhook</AlertDialog.Title>
          <AlertDialog.Description>
            Are you sure you want to delete this webhook
            {webhookDescription ? ` "${webhookDescription}"` : ""}? This action
            cannot be undone.
          </AlertDialog.Description>
        </AlertDialog.Header>
        <AlertDialog.Footer>
          <AlertDialog.Close
            render={
              <Button
                variant="outline"
                color="neutral"
                disabled={deleteWebhookMutation.isPending}
                data-test-id="admin-cancel-delete-webhook"
              >
                Cancel
              </Button>
            }
          />
          <Button
            variant="solid"
            color="danger"
            onClick={handleDelete}
            loading={deleteWebhookMutation.isPending}
            loaderText="Deleting..."
            data-test-id="admin-confirm-delete-webhook"
          >
            Delete
          </Button>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog>
  );
}
