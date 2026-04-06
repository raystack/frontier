'use client';

import { useState } from 'react';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteOrganizationInvitationRequestSchema,
  RemoveOrganizationUserRequestSchema
} from '@raystack/proton/frontier';
import {
  Button,
  Text,
  Dialog,
  Flex,
  toastManager
} from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useTerminology } from '../../../hooks/useTerminology';
import { handleConnectError } from '~/utils/error';

export type RemoveMemberPayload = { memberId: string; invited: string };

export interface RemoveMemberDialogProps {
  handle: ReturnType<typeof Dialog.createHandle<RemoveMemberPayload>>;
  refetch: () => void;
}

export function RemoveMemberDialog({ handle, refetch }: RemoveMemberDialogProps) {
  const { activeOrganization } = useFrontier();
  const organizationId = activeOrganization?.id ?? '';
  const [isLoading, setIsLoading] = useState(false);
  const t = useTerminology();

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      refetch();
    }
  };

  const { mutateAsync: deleteInvitation } = useMutation(
    FrontierServiceQueries.deleteOrganizationInvitation,
    {
      onSuccess: () => {
        handle.close();
        toastManager.add({ title: 'Member deleted', type: 'success' });
      }
    }
  );

  const { mutateAsync: removeUser } = useMutation(
    FrontierServiceQueries.removeOrganizationUser,
    {
      onSuccess: () => {
        handle.close();
        toastManager.add({ title: 'Member deleted', type: 'success' });
      }
    }
  );

  const deleteMember = async (memberId: string, invited: string) => {
    setIsLoading(true);
    try {
      if (invited === 'true') {
        const req = create(DeleteOrganizationInvitationRequestSchema, {
          orgId: organizationId,
          id: memberId
        });
        await deleteInvitation(req);
      } else {
        const req = create(RemoveOrganizationUserRequestSchema, {
          id: organizationId,
          userId: memberId
        });
        await removeUser(req);
      }
    } catch (error) {
      handleConnectError(error, {
        NotFound: (err) => toastManager.add({ title: 'Not found', description: err.message, type: 'error' }),
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as RemoveMemberPayload | undefined;
        return (
          <Dialog.Content width={400}>
            <Dialog.Header>
              <Dialog.Title>Remove member?</Dialog.Title>
            </Dialog.Header>
            <Dialog.Body>
              <Text size="regular" variant="secondary">
                Are you sure you want to remove this member from the{' '}
                {t.organization({ case: 'lower' })}?
              </Text>
            </Dialog.Body>
            <Dialog.Footer>
              <Flex justify="end" gap={5}>
                <Button
                  variant="outline"
                  color="neutral"
                  onClick={() => handle.close()}
                  data-test-id="cancel-remove-member-dialog"
                  disabled={isLoading}
                >
                  Cancel
                </Button>
                <Button
                  variant="solid"
                  color="danger"
                  onClick={() =>
                    payload && deleteMember(payload.memberId, payload.invited)
                  }
                  data-test-id="confirm-remove-member-dialog"
                  disabled={isLoading}
                  loading={isLoading}
                  loaderText="Removing..."
                >
                  Remove
                </Button>
              </Flex>
            </Dialog.Footer>
          </Dialog.Content>
        );
      }}
    </Dialog>
  );
}
