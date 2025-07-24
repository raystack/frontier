import {
  Button,
  toast,
  Image,
  Text,
  Dialog,
  Flex,
  List
} from '@raystack/apsara/v1';
import cross from '~/react/assets/cross.svg';
import { useNavigate } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useState } from 'react';
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
            <Text size="large" weight="medium">
              Chrome on Mac OS x
            </Text>
            <Image
              alt="cross"
              src={cross as unknown as string}
              onClick={() => (isLoading ? null : navigate({ to: '/sessions' }))}
              style={{ cursor: isLoading ? 'not-allowed' : 'pointer' }}
              data-test-id="close-revoke-session-dialog"
            />
          </Flex>
        </Dialog.Header>

        <Dialog.Body className={styles.revokeSessionConfirmBody}>
            <List.Root className={styles.listRoot}>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="100px">Device</List.Label>
                <List.Value>Chrome on Mac OS x</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="100px">IP Address</List.Label>
                <List.Value>203.0.113.25</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="100px">Last Location</List.Label>
                <List.Value>Bangalore, India</List.Value>
              </List.Item>
              <List.Item className={styles.listItem}>
                <List.Label minWidth="100px">Last Active</List.Label>
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
              data-test-id="cancel-revoke-session-dialog"
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="danger"
              onClick={handleRevoke}
              data-test-id="confirm-revoke-session-dialog"
              disabled={isLoading}
            >
              {isLoading ? 'Revoking...' : 'Revoke'}
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
