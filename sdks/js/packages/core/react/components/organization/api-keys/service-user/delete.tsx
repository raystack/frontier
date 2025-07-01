import { useState } from 'react';
import {
  Button,
  Flex,
  Text,
  toast,
  Image,
  Dialog
} from '@raystack/apsara/v1';
import cross from '~/react/assets/cross.svg';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { DEFAULT_API_PLATFORM_APP_NAME } from '~/react/utils/constants';
import styles from './styles.module.css';

export const DeleteServiceAccountKey = () => {
  const { id, tokenId } = useParams({
    from: '/api-keys/$id/key/$tokenId/delete'
  });
  const navigate = useNavigate({ from: '/api-keys/$id/key/$tokenId/delete' });
  const { client, config, activeOrganization } = useFrontier();
  const [isLoading, setIsLoading] = useState(false);

  const orgId = activeOrganization?.id || '';

  async function onDeleteClick() {
    try {
      setIsLoading(true);
      await client?.frontierServiceDeleteServiceUserToken(orgId, id, tokenId);
      navigate({
        to: '/api-keys/$id',
        params: {
          id: id
        },
        state: {
          refetch: true
        }
      });
      toast.success('Service account key revoked');
    } catch (err: any) {
      toast.error('Unable to revoke service account key', {
        description: err?.message
      });
    } finally {
      setIsLoading(false);
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
  const appName = config?.apiPlatform?.appName || DEFAULT_API_PLATFORM_APP_NAME;

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
          <Flex
            direction="column"
            gap={5}
          >
            <Text>
              This is an irreversible action doing this might lead to
              discontinuation of access to the {appName} features. Do you wish
              to proceed?
            </Text>
          </Flex>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex
            justify="end"
            gap={5}
          >
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
              loading={isLoading}
              disabled={isLoading}
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
