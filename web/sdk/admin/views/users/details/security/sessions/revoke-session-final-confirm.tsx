import {
  AlertDialog,
  Button,
  toastManager
} from '@raystack/apsara';
import { handleConnectError } from '~/utils/error';

interface RevokeSessionFinalConfirmProps {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  onConfirm: () => void;
  isLoading?: boolean;
}

export const RevokeSessionFinalConfirm = ({ 
  isOpen, 
  onOpenChange, 
  onConfirm, 
  isLoading = false 
}: RevokeSessionFinalConfirmProps) => {
  const handleConfirm = async () => {
    try {
      onConfirm();
      onOpenChange(false);
    } catch (error) {
      handleConnectError(error, {
        PermissionDenied: () =>
          toastManager.add({
            title: "You don't have permission to perform this action",
            type: 'error',
          }),
        Default: (err) =>
          toastManager.add({
            title: 'Failed to revoke session',
            description: err.rawMessage || 'Something went wrong',
            type: 'error',
          }),
      });
    }
  };

  return (
    <AlertDialog open={isOpen} onOpenChange={onOpenChange}>
      <AlertDialog.Content>
        <AlertDialog.Header>
          <AlertDialog.Title>Revoke</AlertDialog.Title>
          <AlertDialog.Description>
            Are you sure you want to revoke this session? This action cannot be undone.
          </AlertDialog.Description>
        </AlertDialog.Header>

        <AlertDialog.Footer>
          <AlertDialog.Close
            render={
              <Button
                variant="outline"
                color="neutral"
                data-test-id="frontier-ui-cancel-final-revoke-dialog"
                disabled={isLoading}
              >
                Cancel
              </Button>
            }
          />
          <Button
            variant="solid"
            color="danger"
            onClick={handleConfirm}
            data-test-id="frontier-ui-confirm-final-revoke-dialog"
            disabled={isLoading}
            loading={isLoading}
            loaderText="Revoking..."
          >
            Revoke
          </Button>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog>
  );
};
