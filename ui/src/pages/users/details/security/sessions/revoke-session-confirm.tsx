import { useState } from 'react';
import {
  Button,
  toast,
  Dialog,
  Flex,
  List
} from '@raystack/apsara/v1';
import styles from './sessions.module.css';

interface RevokeSessionConfirmProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  sessionInfo?: {
    device: string;
    ipAddress: string;
    location: string;
    lastActive: string;
  };
}

export const RevokeSessionConfirm = ({ isOpen, onOpenChange, sessionInfo }: RevokeSessionConfirmProps) => {
  const [isLoading, setIsLoading] = useState(false);

  const handleRevoke = async () => {
    setIsLoading(true);
    try {
      // TODO: Add API call to revoke session
      toast.success('Session revoked');
      onOpenChange(false);
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '400px', width: '100%' }}
      >
        <Dialog.Header className={styles.revokeSessionConfirmHeader}>
          <Dialog.Title>Safari on Mac OS x</Dialog.Title>
          <Dialog.CloseButton data-test-id="frontier-ui-close-revoke-session-dialog" />
        </Dialog.Header>

        <Dialog.Body className={styles.revokeSessionConfirmBody}>
            <List.Root className={styles.listRoot}>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">Device</List.Label>
                <List.Value>{sessionInfo?.device || "Chrome on Mac OS x"}</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">IP Address</List.Label>
                <List.Value>{sessionInfo?.ipAddress || "203.0.113.25"}</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">Last Location</List.Label>
                <List.Value>{sessionInfo?.location || "Bangalore, India"}</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">Last Active</List.Label>
                <List.Value>{sessionInfo?.lastActive || "10 minutes ago"}</List.Value>
              </List.Item>
            </List.Root>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              onClick={() => onOpenChange(false)}
              data-test-id="frontier-ui-cancel-revoke-session-dialog"
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="danger"
              onClick={handleRevoke}
              data-test-id="frontier-ui-confirm-revoke-session-dialog"
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