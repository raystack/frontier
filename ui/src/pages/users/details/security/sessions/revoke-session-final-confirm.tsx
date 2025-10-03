import {
  Button,
  toast,
  Dialog,
  Flex,
  Text
} from '@raystack/apsara';
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
    } catch (error: any) {
      toast.error('Failed to revoke session', {
        description: error.message || 'Something went wrong'
      });
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '400px', width: '100%' }}
      >
        <Dialog.Header className={styles.revokeSessionConfirmHeader}>
          <Dialog.Title>Revoke</Dialog.Title>
          <Dialog.CloseButton data-test-id="frontier-ui-close-final-revoke-dialog" />
        </Dialog.Header>

        <Dialog.Body className={styles.revokeSessionFinalConfirmBody}>
          <Flex direction="column" gap={4}>
            <Text size="small" variant="secondary">
              Are you sure you want to revoke this session? This action cannot be undone.
            </Text>
          </Flex>
        </Dialog.Body>

        <Dialog.Footer>
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
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
