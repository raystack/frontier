'use client';

import { useState } from 'react';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteServiceUserTokenRequestSchema
} from '@raystack/proton/frontier';
import {
  Button,
  Text,
  AlertDialog,
  Flex,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useTerminology } from '../../../hooks/useTerminology';
import styles from './revoke-token-dialog.module.css';

export type RevokeTokenPayload = { tokenId: string };

export interface RevokeTokenDialogProps {
  handle: ReturnType<typeof AlertDialog.createHandle<RevokeTokenPayload>>;
  serviceUserId: string;
  onRevoked: (tokenId: string) => void;
}

export function RevokeTokenDialog({
  handle,
  serviceUserId,
  onRevoked
}: RevokeTokenDialogProps) {
  const { activeOrganization } = useFrontier();
  const orgId = activeOrganization?.id ?? '';
  const t = useTerminology();
  const [isLoading, setIsLoading] = useState(false);

  const { mutateAsync: deleteServiceUserToken } = useMutation(
    FrontierServiceQueries.deleteServiceUserToken
  );

  const handleRevoke = async (tokenId: string) => {
    setIsLoading(true);
    try {
      await deleteServiceUserToken(
        create(DeleteServiceUserTokenRequestSchema, {
          id: serviceUserId,
          tokenId,
          orgId
        })
      );
      onRevoked(tokenId);
      handle.close();
      toastManager.add({ title: 'Service account key revoked', type: 'success' });
    } catch (error: unknown) {
      toastManager.add({
        title: 'Unable to revoke service account key',
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
        const payload = rawPayload as RevokeTokenPayload | undefined;
        return (
          <AlertDialog.Content width={400} showCloseButton={false}>
            <AlertDialog.Body className={styles.body}>
              <AlertDialog.Title>Revoke API Key</AlertDialog.Title>
              <Text size="small" variant="secondary">
                This is an irreversible action doing this might lead to
                discontinuation of access to the {t.appName()} features. Do you
                wish to proceed?
              </Text>
            </AlertDialog.Body>
            <AlertDialog.Footer>
              <Flex justify="end" gap={5}>
                <Button
                  variant="outline"
                  color="neutral"
                  onClick={() => handle.close()}
                  disabled={isLoading}
                  data-test-id="frontier-sdk-revoke-token-cancel-btn"
                >
                  Cancel
                </Button>
                <Button
                  variant="solid"
                  color="danger"
                  onClick={() => payload && handleRevoke(payload.tokenId)}
                  disabled={isLoading}
                  loading={isLoading}
                  loaderText="Revoking..."
                  data-test-id="frontier-sdk-revoke-token-confirm-btn"
                >
                  Revoke
                </Button>
              </Flex>
            </AlertDialog.Footer>
          </AlertDialog.Content>
        );
      }}
    </AlertDialog>
  );
}
