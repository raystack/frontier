import { Button, Flex, Text, toast, Image, Dialog } from '@raystack/apsara';
import cross from '~/react/assets/cross.svg';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import styles from './styles.module.css';
import { useTerminology } from '~/react/hooks/useTerminology';
import { useQueryClient } from '@tanstack/react-query';
import { useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  ListServiceUserTokensRequestSchema,
  DeleteServiceUserTokenRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

export const DeleteServiceAccountKey = () => {
  const { id, tokenId } = useParams({
    from: '/api-keys/$id/key/$tokenId/delete'
  });
  const navigate = useNavigate({ from: '/api-keys/$id/key/$tokenId/delete' });
  const { activeOrganization } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const orgId = activeOrganization?.id || '';

  const { mutateAsync: deleteServiceUserToken, isPending } = useMutation(
    FrontierServiceQueries.deleteServiceUserToken
  );

  async function onDeleteClick() {
    try {
      await deleteServiceUserToken(
        create(DeleteServiceUserTokenRequestSchema, {
          id,
          tokenId,
          orgId
        })
      );

      // Invalidate service user tokens query
      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: FrontierServiceQueries.listServiceUserTokens,
          transport,
          input: create(ListServiceUserTokensRequestSchema, {
            id,
            orgId
          }),
          cardinality: 'finite'
        })
      });

      navigate({
        to: '/api-keys/$id',
        params: {
          id: id
        }
      });
      toast.success('Service account key revoked');
    } catch (error: unknown) {
      toast.error('Unable to revoke service account key', {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  }

  function onCancel() {
    navigate({
      to: '/api-keys/$id',
      params: {
        id: id
      }
    });
  }

  const t = useTerminology();

  return (
    <Dialog open={true}>
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
              onClick={onCancel}
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
              onClick={onCancel}
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
