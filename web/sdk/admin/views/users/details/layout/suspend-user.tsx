import { useState } from "react";
import { AlertDialog, Button, Flex, Text, toastManager } from "@raystack/apsara";
import { useTerminology } from "../../../../hooks/useTerminology";

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
  const t = useTerminology();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSuspend = async () => {
    try {
      setIsSubmitting(true);
      toastManager.add({ title: `${t.user({ case: "capital" })} suspended successfully`, type: "success" });
      onSubmit?.();
    } catch (error) {
      console.error(error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <AlertDialog open onOpenChange={onClose}>
      <AlertDialog.Content>
        <AlertDialog.Body>
          <Flex direction="column" gap={3}>
            <AlertDialog.Title>Suspend {t.user({ case: "capital" })}</AlertDialog.Title>
            <Text variant="secondary">
              Suspending this {t.user({ case: "lower" })} will permanently restrict access to its
              content, disable communication, and prevent any future
              interactions. Are you sure you want to proceed?
            </Text>
          </Flex>
        </AlertDialog.Body>
        <AlertDialog.Footer>
          <Button
            type="button"
            variant="outline"
            color="neutral"
            onClick={onClose}
            disabled={isSubmitting}
            data-test-id="admin-user-details-suspend-cancel"
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="solid"
            color="danger"
            data-test-id="admin-user-details-suspend-confirm"
            onClick={handleSuspend}
            loading={isSubmitting}
            loaderText="Suspending..."
          >
            Suspend
          </Button>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog>
  );
};
