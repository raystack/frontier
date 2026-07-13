import {
  AlertDialog,
  Button,
  toastManager,
  Flex,
  Text
} from '@raystack/apsara';
import { handleConnectError } from '~/utils/error';
import styles from './sessions.module.css';

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
      <AlertDialog.Content
        style={{ padding: 0, maxWidth: '400px', width: '100%' }}
      >
        <AlertDialog.Header className={styles.revokeSessionConfirmHeader}>
          <AlertDialog.Title>Revoke</AlertDialog.Title>
        </AlertDialog.Header>

        <AlertDialog.Body className={styles.revokeSessionFinalConfirmBody}>
          <Flex direction="column" gap={4}>
            <Text size="small" variant="secondary">
              Are you sure you want to revoke this session? This action cannot be undone.
            </Text>
          </Flex>
        </AlertDialog.Body>

        <AlertDialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              onClick={() => onOpenChange(false)}
              data-test-id="frontier-ui-cancel-final-revoke-dialog"
              disabled={isLoading}
            >
              Cancel
            </Button>
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
          </Flex>
        </AlertDialog.Footer>
      </AlertDialog.Content>
    </AlertDialog>
  );
};
