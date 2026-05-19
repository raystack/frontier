import { useState, useMemo } from 'react';
import { useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';
import { AlertDialog, Button, Flex, Skeleton, Text } from '@raystack/apsara-v1';
import { useSessions } from '~/react/hooks/useSessionsV1';
import { RevokeSessionConfirmDialog } from './revoke-session-confirm-dialog';
import styles from './revoke-session-dialog.module.css';

export interface RevokeSessionDialogProps {
  open: boolean;
  onOpenChange: (value: boolean) => void;
  sessionId: string;
  onLogout?: () => void;
  onRevoke?: () => void;
}

export const RevokeSessionDialog = ({
  open,
  onOpenChange,
  sessionId,
  onLogout,
  onRevoke
}: RevokeSessionDialogProps) => {
  const { sessions, revokeSession, isRevokingSession } = useSessions();
  const [isConfirmOpen, setIsConfirmOpen] = useState(false);

  const handleClose = () => onOpenChange(false);

  const { mutate: logout } = useMutation(FrontierServiceQueries.authLogout, {
    onSuccess: () => {
      onLogout?.();
    },
    onError: () => {
      if (sessionId) {
        revokeSession(sessionId);
        handleClose();
        setIsConfirmOpen(false);
      }
    }
  });

  const sessionData = useMemo(() => {
    if (!sessionId || sessions.length === 0) return null;
    return sessions.find(s => s.id === sessionId) ?? null;
  }, [sessionId, sessions]);

  const detailRows = [
    { label: 'Device', value: `${sessionData?.browser} on ${sessionData?.operatingSystem}` },
    { label: 'IP address', value: sessionData?.ipAddress },
    { label: 'Last location', value: sessionData?.location },
    { label: 'Last active', value: sessionData?.isCurrent ? 'Current session' : sessionData?.lastActive }
  ];

  const handleActionClick = () => {
    setIsConfirmOpen(true);
  };

  const handleFinalConfirm = () => {
    if (!sessionId) return;
    if (sessionData?.isCurrent) {
      logout({});
      return;
    }
    return revokeSession(sessionId, {
      onSuccess: () => {
        onRevoke?.();
        handleClose();
        setIsConfirmOpen(false);
      },
    })
  };

  return (
    <>
      <AlertDialog open={open} onOpenChange={onOpenChange}>
        <AlertDialog.Content>
          <AlertDialog.Header>
            {sessionData ? (
              <AlertDialog.Title>
                {sessionData.browser} on {sessionData.operatingSystem}
              </AlertDialog.Title>
            ) : (
              <Skeleton height="24px" width="180px" />
            )}
          </AlertDialog.Header>

          <AlertDialog.Body>
            <Flex direction="column" >
              {detailRows.map(({ label, value }) => (
                <Flex
                  key={label}
                  gap={7}
                  align="center"
                  className={styles.detailRow}
                >
                  {sessionData ?
                    <>
                      <Text size="small" weight="medium" variant="secondary" className={styles.detailLabel}>
                        {label}
                      </Text>
                      <Text size="small" weight="medium">{value}</Text>
                    </>
                    : <Skeleton height="16px" />}
                </Flex>
              ))}
            </Flex>
          </AlertDialog.Body>

          <AlertDialog.Footer>
            <Button
              variant="outline"
              color="neutral"
              onClick={handleClose}
              data-test-id="frontier-sdk-cancel-revoke-session-dialog"
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="danger"
              onClick={handleActionClick}
              data-test-id="frontier-sdk-confirm-revoke-session-dialog"
            >
              {sessionData?.isCurrent ? 'Log out' : 'Revoke'}
            </Button>
          </AlertDialog.Footer>
        </AlertDialog.Content>
      </AlertDialog>

      <RevokeSessionConfirmDialog
        open={isConfirmOpen}
        onOpenChange={setIsConfirmOpen}
        onConfirm={handleFinalConfirm}
        isLoading={isRevokingSession}
        isCurrentSession={sessionData?.isCurrent}
      />
    </>
  );
};
