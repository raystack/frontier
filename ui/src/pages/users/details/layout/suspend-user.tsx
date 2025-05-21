import { useState } from "react";
import { Button, Dialog, Flex, Text, toast } from "@raystack/apsara/v1";
import { api } from "~/api";

interface SuspendDropdownProps {
  userId: string;
  onClose: () => void;
  onSubmit?: () => void;
}

export const SuspendUser = ({
  userId,
  onClose,
  onSubmit,
}: SuspendDropdownProps) => {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSuspend = async () => {
    try {
      setIsSubmitting(true);
      await api.frontierServiceDisableUser(userId, {});
      toast.success("User suspended successfully");
      onSubmit?.();
    } catch (error) {
      console.error(error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content width={400}>
        <Dialog.Body>
          <Flex direction="column" gap="small">
            <Dialog.Title>Suspend User</Dialog.Title>
            <Text variant="secondary">
              Suspending this user will permanently restrict access to its
              content, disable communication, and prevent any future
              interactions. Are you sure you want to proceed?
            </Text>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer>
          <Dialog.Close asChild>
            <Button
              type="button"
              variant="outline"
              color="neutral"
              data-test-id="admin-ui-user-details-suspend-cancel">
              Cancel
            </Button>
          </Dialog.Close>
          <Button
            type="button"
            variant="solid"
            color="danger"
            data-test-id="admin-ui-user-details-suspend-confirm"
            onClick={handleSuspend}
            loading={isSubmitting}
            loaderText="Suspending...">
            Suspend
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
