'use client';

import { useState } from 'react';
import { create } from '@bufbuild/protobuf';
import { useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteServiceUserRequestSchema,
  ListOrganizationServiceUsersRequestSchema
} from '@raystack/proton/frontier';
import {
  Button,
  Text,
  AlertDialog,
  Flex,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useQueryClient } from '@tanstack/react-query';
import styles from './delete-service-account-dialog.module.css';

export type DeleteServiceAccountPayload = { serviceAccountId: string };

export interface DeleteServiceAccountDialogProps {
  handle: ReturnType<typeof AlertDialog.createHandle<DeleteServiceAccountPayload>>;
  refetch: () => void;
}

export function DeleteServiceAccountDialog({ handle, refetch }: DeleteServiceAccountDialogProps) {
  const { activeOrganization } = useFrontier();
  const orgId = activeOrganization?.id ?? '';
  const queryClient = useQueryClient();
  const transport = useTransport();
  const [isLoading, setIsLoading] = useState(false);

  const { mutateAsync: deleteServiceUser } = useMutation(
    FrontierServiceQueries.deleteServiceUser
  );

  const handleDelete = async (serviceAccountId: string) => {
    setIsLoading(true);
    try {
      await deleteServiceUser(
        create(DeleteServiceUserRequestSchema, {
          id: serviceAccountId,
          orgId
        })
      );

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

      handle.close();
      refetch();
      toastManager.add({ title: 'Service account deleted', type: 'success' });
    } catch (error: unknown) {
      toastManager.add({
        title: 'Unable to delete service account',
        description: error instanceof Error ? error.message : 'Unknown error',
        type: 'error'
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AlertDialog handle={handle}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as DeleteServiceAccountPayload | undefined;
        return (
          <AlertDialog.Content width={400} showCloseButton={false}>
            <AlertDialog.Body className={styles.body}>
              <AlertDialog.Title>Delete Service Account</AlertDialog.Title>
              <Text size="small" variant="secondary">
                This action is irreversible and may result in the deletion of all
                keys associated with this account. Are you sure you want to
                proceed?
              </Text>
            </AlertDialog.Body>
            <AlertDialog.Footer>
              <Flex justify="end" gap={5}>
                <Button
                  variant="outline"
                  color="neutral"
                  onClick={() => handle.close()}
                  data-test-id="frontier-sdk-delete-service-account-cancel-btn"
                  disabled={isLoading}
                >
                  Cancel
                </Button>
                <Button
                  variant="solid"
                  color="danger"
                  onClick={() =>
                    payload && handleDelete(payload.serviceAccountId)
                  }
                  data-test-id="frontier-sdk-delete-service-account-confirm-btn"
                  disabled={isLoading}
                  loading={isLoading}
                  loaderText="Deleting..."
                >
                  Delete
                </Button>
              </Flex>
            </AlertDialog.Footer>
          </AlertDialog.Content>
        );
      }}
    </AlertDialog>
  );
}
