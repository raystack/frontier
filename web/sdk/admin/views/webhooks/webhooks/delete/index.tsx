import { Button, Dialog, Flex, Text } from "@raystack/apsara";
import { toast } from "sonner";
import type { useMutation } from "@connectrpc/connect-query";

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
      toast.success("Webhook deleted");
      onOpenChange(false);
    } catch (err) {
      console.error("Failed to delete webhook:", err);
      toast.error("Failed to delete webhook");
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%' }}
      >
        <Flex direction="column" gap="large" style={{ padding: "24px" }}>
          <Flex direction="column" gap="medium">
            <Text size={5} weight={500}>
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
