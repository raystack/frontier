import { useState } from 'react';
import { Dialog } from '@raystack/apsara';
import { Button, Flex, Text, toast, Separator, Image } from '@raystack/apsara/v1';
import cross from '~/react/assets/cross.svg';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import styles from './styles.module.css';

export const DeleteServiceAccount = () => {
  const { id } = useParams({ from: '/api-keys/$id/delete' });
  const navigate = useNavigate({ from: '/api-keys/$id/delete' });
  const { client, activeOrganization: organization } = useFrontier();
  const [isLoading, setIsLoading] = useState(false);

  const orgId = organization?.id || '';

  async function onDeleteClick() {
    try {
      setIsLoading(true);
      await client?.frontierServiceDeleteServiceUser(orgId, id);
      navigate({
        to: '/api-keys',
        state: {
          refetch: true
        }
      });
      toast.success('Service account deleted');
    } catch (err: any) {
      toast.error('Unable to delete service account', {
        description: err?.message
      });
    } finally {
      setIsLoading(false);
    }
  }

  function onCancel() {
    navigate({ to: '/api-keys' });
  }

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        overlayClassname={styles.overlay}
        className={styles.addDialogContent}
      >
        <Flex justify="between" className={styles.addDialogForm}>
          <Text size={6} weight="medium">
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
        <Separator />

        <Flex
          direction="column"
          gap="medium"
          className={styles.addDialogFormContent}
        >
          <Text>
            This is an irreversible and permanent action doing this might result
            in deletion of the service account and the keys associated with it.
            Do you wish to proceed?
          </Text>
        </Flex>
        <Separator />
        <Flex
          justify="end"
          className={styles.addDialogFormBtnWrapper}
          gap={'medium'}
        >
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
            loading={isLoading}
            disabled={isLoading}
            onClick={onDeleteClick}
            loaderText="Deleting..."
          >
            I Understand and Delete
          </Button>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
};
