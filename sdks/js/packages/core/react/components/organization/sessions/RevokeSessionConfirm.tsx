import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import {
  Button,
  toast,
  Text,
  Dialog,
  Flex,
  List,
  IconButton,
  Image
} from '@raystack/apsara/v1';
import { useFrontier } from '~/react/contexts/FrontierContext';
import cross from '~/react/assets/cross.svg';
import styles from './sessions.module.css';

export const RevokeSessionConfirm = () => {
  const navigate = useNavigate({ from: '/sessions/revoke' });
  const { client, activeOrganization } = useFrontier();
  const [isLoading, setIsLoading] = useState(false);

  const handleRevoke = async () => {
    setIsLoading(true);
    try {
      // TODO: Add API call to revoke session
      navigate({ to: '/sessions' });
      toast.success('Session revoked');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={true} onOpenChange={() => navigate({ to: '/sessions' })}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '400px', width: '100%' }}
      >
        <Dialog.Header className={styles.revokeSessionConfirmHeader}>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="regular">
              Chrome on Mac OS x
            </Text>
            <IconButton
              size={3}
              onClick={() => !isLoading && navigate({ to: '/sessions' })}
              disabled={isLoading}
              data-test-id="frontier-sdk-close-revoke-session-dialog"
              aria-label="Close dialog"
            >
              <Image
                alt="close"
                src={cross as unknown as string}
              />
            </IconButton>
          </Flex>
        </Dialog.Header>

        <Dialog.Body className={styles.revokeSessionConfirmBody}>
            <List.Root className={styles.listRoot}>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">Device</List.Label>
                <List.Value>Chrome on Mac OS x</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">IP Address</List.Label>
                <List.Value>203.0.113.25</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">Last Location</List.Label>
                <List.Value>Bangalore, India</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="120px">Last Active</List.Label>
                <List.Value>10 minutes ago</List.Value>
              </List.Item>
            </List.Root>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              onClick={() => navigate({ to: '/sessions' })}
              data-test-id="frontier-sdk-cancel-revoke-session-dialog"
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="danger"
              onClick={handleRevoke}
              data-test-id="frontier-sdk-confirm-revoke-session-dialog"
              disabled={isLoading}
              loading={isLoading}
              loadingText="Revoking..."
            >
              Revoke
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
