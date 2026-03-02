import { Button, Flex, Text, toast, Image, Dialog } from '@raystack/apsara';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import styles from './service-user.module.css';
import { useTerminology } from '~/react/hooks/useTerminology';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteServiceUserTokenRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { useServiceUserTokens } from '../hooks/useServiceUserTokens';

export interface DeleteServiceUserKeyDialogProps {
  open: boolean;
  onOpenChange?: (value: boolean) => void;
  serviceUserId: string;
  tokenId: string;
}

export const DeleteServiceUserKeyDialog = ({
  open,
  onOpenChange,
  serviceUserId,
  tokenId
}: DeleteServiceUserKeyDialogProps) => {
  const { activeOrganization } = useFrontier();

  const orgId = activeOrganization?.id || '';

  const { removeToken } = useServiceUserTokens({
    id: serviceUserId,
    orgId,
    enableFetch: false
  });

  const { mutateAsync: deleteServiceUserToken, isPending } = useMutation(
    FrontierServiceQueries.deleteServiceUserToken
  );

  const handleClose = () => onOpenChange?.(false);

  async function onDeleteClick() {
    try {
      await deleteServiceUserToken(
        create(DeleteServiceUserTokenRequestSchema, {
          id: serviceUserId,
          tokenId,
          orgId
        })
      );

      // Remove token from cache
      removeToken(tokenId);

      handleClose();
      toast.success('Service account key revoked');
    } catch (error: unknown) {
      toast.error('Unable to revoke service account key', {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  }

  const t = useTerminology();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Content
        overlayClassName={styles.overlay}
        className={styles.addDialogContent}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Revoke API Key
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={handleClose}
              data-test-id="frontier-sdk-revoke-service-account-key-close-btn"
            />
          </Flex>
        </Dialog.Header>

        <Dialog.Body>
          <Flex direction="column" gap={5}>
            <Text>
              This is an irreversible action doing this might lead to
              discontinuation of access to the {t.appName()} features. Do you
              wish to proceed?
            </Text>
          </Flex>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              size="normal"
              data-test-id="frontier-sdk-revoke-service-account-key-cancel-btn"
              onClick={handleClose}
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="danger"
              size="normal"
              data-test-id="frontier-sdk-revoke-service-account-key-confirm-btn"
              loading={isPending}
              disabled={isPending}
              onClick={onDeleteClick}
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
