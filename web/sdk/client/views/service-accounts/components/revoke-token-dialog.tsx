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
  AlertDialog,
  Flex,
  toastManager
} from '@raystack/apsara';
import { useFrontier } from '~/client/contexts/FrontierContext';
import { useTerminology } from '~/client/hooks/useTerminology';

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
          <AlertDialog.Content>
            <AlertDialog.Header>
              <AlertDialog.Title>Revoke API Key</AlertDialog.Title>
              <AlertDialog.Description>
                This is an irreversible action doing this might lead to
                discontinuation of access to the {t.appName()} features. Do you
                wish to proceed?
              </AlertDialog.Description>
            </AlertDialog.Header>
            <AlertDialog.Footer>
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
            </AlertDialog.Footer>
          </AlertDialog.Content>
        );
      }}
    </AlertDialog>
  );
}
