import { useState } from 'react';
import {
  Button,
  toast,
  Dialog,
  Flex,
  List
} from '@raystack/apsara';
import { RevokeSessionFinalConfirm } from './revoke-session-final-confirm';
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
  onRevokeConfirm: () => void;
}

export const RevokeSessionConfirm = ({ isOpen, onOpenChange, sessionInfo, onRevokeConfirm }: RevokeSessionConfirmProps) => {
  const [isFinalConfirmOpen, setIsFinalConfirmOpen] = useState(false);

  const handleRevoke = () => {
    setIsFinalConfirmOpen(true);
  };

  const handleFinalConfirm = async () => {
    try {
      onRevokeConfirm();
      onOpenChange(false);
      toast.success('Session revoked successfully');
    } catch (error: any) {
      toast.error('Failed to revoke session', {
        description: error.message || 'Something went wrong'
      });
    }
  };

  return (
    <>
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '400px', width: '100%' }}
      >
        <Dialog.Header className={styles.revokeSessionConfirmHeader}>
          <Dialog.Title>{sessionInfo?.device || "Unknown Device"}</Dialog.Title>
          <Dialog.CloseButton data-test-id="frontier-ui-close-revoke-session-dialog" />
        </Dialog.Header>

        <Dialog.Body className={styles.revokeSessionConfirmBody}>
            <List className={styles.listRoot}>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">Device</List.Label>
                <List.Value>{sessionInfo?.device || "Unknown"}</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">IP Address</List.Label>
                <List.Value>{sessionInfo?.ipAddress || "Unknown"}</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">Last Location</List.Label>
                <List.Value>{sessionInfo?.location || "Unknown"}</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">Last Active</List.Label>
                <List.Value>{sessionInfo?.lastActive || "Unknown"}</List.Value>
              </List.Item>
            </List>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              onClick={() => onOpenChange(false)}
              data-test-id="frontier-ui-cancel-revoke-session-dialog"
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="danger"
              onClick={handleRevoke}
              data-test-id="frontier-ui-confirm-revoke-session-dialog"
            >
              Revoke
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>

    <RevokeSessionFinalConfirm
      isOpen={isFinalConfirmOpen}
      onOpenChange={setIsFinalConfirmOpen}
      onConfirm={handleFinalConfirm}
    />
    </>
  );
}; 