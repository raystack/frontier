'use client';

import { useState } from 'react';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteCurrentUserPATRequestSchema
} from '@raystack/proton/frontier';
import { AlertDialog, Button, toastManager } from '@raystack/apsara-v1';
import styles from './revoke-pat-dialog.module.css';
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
        <AlertDialog.Content width={400} showCloseButton={false}>
          <AlertDialog.Body className={styles.body}>
            <AlertDialog.Title>Revoke</AlertDialog.Title>
            <AlertDialog.Description>
              This action cannot be undone. Revoking this token will
              permanently remove access for any users using it. You&apos;ll
              need to generate a new token if access is required again.
            </AlertDialog.Description>
          </AlertDialog.Body>
          <AlertDialog.Footer justify="end" gap={5}>
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
          </AlertDialog.Footer>
        </AlertDialog.Content>
      )}
    </AlertDialog>
  );
}
