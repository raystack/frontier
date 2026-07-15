import { useState } from 'react';
import {
  Dialog,
  Button,
  List
} from '@raystack/apsara';
import { RevokeSessionFinalConfirm } from './revoke-session-final-confirm';
import { formatDeviceDisplay } from './index';
import styles from './sessions.module.css';

interface RevokeSessionConfirmProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  sessionInfo?: {
    browser: string;
    operatingSystem: string;
    ipAddress: string;
    location: string;
    lastActive: string;
  };
  onRevokeConfirm: () => void;
  isLoading?: boolean;
}

export const RevokeSessionConfirm = ({ isOpen, onOpenChange, sessionInfo, onRevokeConfirm, isLoading = false }: RevokeSessionConfirmProps) => {
  const [isFinalConfirmOpen, setIsFinalConfirmOpen] = useState(false);

  const handleRevoke = () => {
    setIsFinalConfirmOpen(true);
  };

  const handleFinalConfirm = () => {
    onRevokeConfirm();
    onOpenChange(false);
  };

  return (
    <>
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Content showCloseButton={false}>
        <Dialog.Header className={styles.revokeSessionConfirmHeader}>
          <Dialog.Title>
            {sessionInfo ? formatDeviceDisplay(sessionInfo.browser, sessionInfo.operatingSystem) : "Unknown browser and OS"}
          </Dialog.Title>
        </Dialog.Header>

        <Dialog.Body className={styles.revokeSessionConfirmBody}>
            <List className={styles.listRoot}>
              <List.Item className={styles.listItem}>
                <List.Label className={styles.listLabel}>Device</List.Label>
                <List.Value>{sessionInfo ? formatDeviceDisplay(sessionInfo.browser, sessionInfo.operatingSystem) : "Unknown"}</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label className={styles.listLabel}>IP Address</List.Label>
                <List.Value>{sessionInfo?.ipAddress || "Unknown"}</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label className={styles.listLabel}>Last Location</List.Label>
                <List.Value>{sessionInfo?.location || "Unknown"}</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label className={styles.listLabel}>Last Active</List.Label>
                <List.Value>{sessionInfo?.lastActive || "Unknown"}</List.Value>
              </List.Item>
            </List>
        </Dialog.Body>

        <Dialog.Footer>
          <Dialog.Close
            render={
              <Button
                variant="outline"
                color="neutral"
                disabled={isLoading}
                data-test-id="frontier-ui-cancel-revoke-session-dialog"
              >
                Cancel
              </Button>
            }
          />
          <Button
            variant="solid"
            color="danger"
            onClick={handleRevoke}
            loading={isLoading}
            loaderText="Revoking..."
            data-test-id="frontier-ui-confirm-revoke-session-dialog"
          >
            Revoke
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>

    <RevokeSessionFinalConfirm
      isOpen={isFinalConfirmOpen}
      onOpenChange={setIsFinalConfirmOpen}
      onConfirm={handleFinalConfirm}
      isLoading={isLoading}
    />
    </>
  );
}; 