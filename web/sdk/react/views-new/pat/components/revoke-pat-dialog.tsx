'use client';

import { useState } from 'react';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteCurrentUserPATRequestSchema
} from '@raystack/proton/frontier';
import {
  AlertDialog,
  Button,
  Flex,
  Text,
  toastManager
} from '@raystack/apsara-v1';
import { handleConnectError } from '~/utils/error';

export interface RevokePATDialogProps {
  handle: ReturnType<typeof AlertDialog.createHandle<string>>;
  onRevoked?: () => void;
}

export function RevokePATDialog({ handle, onRevoked }: RevokePATDialogProps) {
  const [isLoading, setIsLoading] = useState(false);

  const { mutateAsync: deletePAT } = useMutation(
    FrontierServiceQueries.deleteCurrentUserPAT
  );

  const handleRevoke = async (patId: string) => {
    setIsLoading(true);
    try {
      await deletePAT(
        create(DeleteCurrentUserPATRequestSchema, { id: patId })
      );
      handle.close();
      toastManager.add({ title: 'Token revoked', type: 'success' });
      onRevoked?.();
    } catch (error) {
      handleConnectError(error, {
        Default: err =>
          toastManager.add({
            title: 'Something went wrong',
            description: err.message,
            type: 'error'
          })
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AlertDialog handle={handle}>
      {({ payload: patId }) => (
        <AlertDialog.Content width={400}>
          <AlertDialog.Header>
            <AlertDialog.Title>Revoke</AlertDialog.Title>
          </AlertDialog.Header>
          <AlertDialog.Body>
            <Text size="small" variant="secondary">
              This is an irreversible action, doing this might lead to permanent
              deletion of the data. Do you wish to proceed?
            </Text>
          </AlertDialog.Body>
          <AlertDialog.Footer>
            <Flex justify="end" gap={5}>
              <Button
                variant="outline"
                color="neutral"
                onClick={() => handle.close()}
                disabled={isLoading}
                data-test-id="frontier-sdk-revoke-pat-cancel-btn"
              >
                Cancel
              </Button>
              <Button
                variant="solid"
                color="danger"
                onClick={() => patId && handleRevoke(patId)}
                disabled={isLoading}
                loading={isLoading}
                loaderText="Revoking..."
                data-test-id="frontier-sdk-revoke-pat-confirm-btn"
              >
                Revoke
              </Button>
            </Flex>
          </AlertDialog.Footer>
        </AlertDialog.Content>
      )}
    </AlertDialog>
  );
}
