import { Button, Dialog, Flex, Text, toastManager } from "@raystack/apsara-v1";
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
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
      >
        <Flex direction="column" gap={9} style={{ padding: "24px" }}>
          <Flex direction="column" gap={5}>
            <Text size="large" weight="medium">
              Delete Webhook
            </Text>
            <Text>
              Are you sure you want to delete this webhook
              {webhookDescription ? ` "${webhookDescription}"` : ""}? This action
              cannot be undone.
            </Text>
          </Flex>

          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              onClick={() => onOpenChange(false)}
              disabled={deleteWebhookMutation.isPending}
              data-test-id="admin-cancel-delete-webhook"
            >
              Cancel
            </Button>
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
          </Flex>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}
