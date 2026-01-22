import { Button, Flex, Text, toast, Image, Dialog } from '@raystack/apsara';
import cross from '~/react/assets/cross.svg';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import styles from './styles.module.css';
import { useQueryClient } from '@tanstack/react-query';
import { useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  ListOrganizationServiceUsersRequestSchema,
  DeleteServiceUserRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

export const DeleteServiceAccount = () => {
  const { id } = useParams({ from: '/api-keys/$id/delete' });
  const navigate = useNavigate({ from: '/api-keys/$id/delete' });
  const { activeOrganization: organization } = useFrontier();
  const queryClient = useQueryClient();
  const transport = useTransport();

  const orgId = organization?.id || '';

  const { mutateAsync: deleteServiceUser, isPending } = useMutation(
    FrontierServiceQueries.deleteServiceUser
  );

  async function onDeleteClick() {
    try {
      await deleteServiceUser(
        create(DeleteServiceUserRequestSchema, {
          id,
          orgId
        })
      );

      // Invalidate service users query
      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          schema: FrontierServiceQueries.listOrganizationServiceUsers,
          transport,
          input: create(ListOrganizationServiceUsersRequestSchema, {
            id: orgId
          }),
          cardinality: 'finite'
        })
      });

      navigate({ to: '/api-keys' });
      toast.success('Service account deleted');
    } catch (error: unknown) {
      toast.error('Unable to delete service account', {
        description: error instanceof Error ? error.message : 'Unknown error'
      });
    }
  }

  function onCancel() {
    navigate({ to: '/api-keys' });
  }

  return (
    <Dialog open={true}>
      <Dialog.Content
        overlayClassName={styles.overlay}
        className={styles.addDialogContent}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Delete Service Account
            </Text>

            <Image
              alt="cross"
              style={{ cursor: 'pointer' }}
              src={cross as unknown as string}
              onClick={() => navigate({ to: '/api-keys' })}
              data-test-id="frontier-sdk-delete-service-account-close-btn"
            />
          </Flex>
        </Dialog.Header>

        <Dialog.Body>
          <Flex direction="column" gap={5}>
            <Text>
              This is an irreversible and permanent action doing this might
              result in deletion of the service account and the keys associated
              with it. Do you wish to proceed?
            </Text>
          </Flex>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              size="normal"
              data-test-id="frontier-sdk-delete-service-account-cancel-btn"
              onClick={onCancel}
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="danger"
              size="normal"
              data-test-id="frontier-sdk-delete-service-account-confirm-btn"
              loading={isPending}
              disabled={isPending}
              onClick={onDeleteClick}
              loaderText="Deleting..."
            >
              I Understand and Delete
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
