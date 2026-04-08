import { AlertDialog, Button } from '@raystack/apsara-v1';

interface RevokeSessionConfirmDialogProps {
  open: boolean;
  onOpenChange: (isOpen: boolean) => void;
  onConfirm: () => void;
  isLoading?: boolean;
  isCurrentSession?: boolean;
}

export const RevokeSessionConfirmDialog = ({
  open,
  onOpenChange,
  onConfirm,
  isLoading = false,
  isCurrentSession = false
}: RevokeSessionConfirmDialogProps) => {
  const handleConfirm = () => {
    onConfirm();
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialog.Content width={400} showCloseButton={false}>
        <AlertDialog.Body>
          <AlertDialog.Title>
            {isCurrentSession ? 'Log out' : 'Revoke'}
          </AlertDialog.Title>
          <AlertDialog.Description>
            Are you sure you want to {isCurrentSession ? 'log out' : 'revoke'} of this session? This action cannot be undone.
          </AlertDialog.Description>
        </AlertDialog.Body>
        <AlertDialog.Footer justify="end" gap={5}>
          <Button
            variant="outline"
            color="neutral"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
            data-test-id="frontier-sdk-cancel-final-revoke-dialog"
          >
            Cancel
          </Button>
          <Button
            variant="solid"
            color="danger"
            onClick={handleConfirm}
            disabled={isLoading}
            loading={isLoading}
            loaderText={isCurrentSession ? 'Signing out...' : 'Revoking...'}
            data-test-id="frontier-sdk-confirm-final-revoke-dialog"
          >
            {isCurrentSession ? 'Log out' : 'Revoke'}
          </Button>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog>
  );
};
