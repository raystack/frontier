'use client';

import { forwardRef, useImperativeHandle, useState } from 'react';
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

export interface RemoveMemberDialogHandle {
  open: (memberId: string, invited: string) => void;
}

export interface RemoveMemberDialogProps {
  refetch: () => void;
}

export const RemoveMemberDialog = forwardRef<
  RemoveMemberDialogHandle,
  RemoveMemberDialogProps
>(function RemoveMemberDialog({ refetch }, ref) {
  const { activeOrganization } = useFrontier();
  const organizationId = activeOrganization?.id ?? '';
  const [state, setState] = useState<{
    open: boolean;
    memberId: string;
    invited: string;
  }>({ open: false, memberId: '', invited: 'false' });
  const [isLoading, setIsLoading] = useState(false);
  const t = useTerminology();

  useImperativeHandle(ref, () => ({
    open: (memberId: string, invited: string) =>
      setState({ open: true, memberId, invited })
  }));

  const handleOpenChange = (value: boolean) => {
    if (!value) {
      setState({ open: false, memberId: '', invited: 'false' });
      refetch();
    }
  };

  const { mutateAsync: deleteInvitation } = useMutation(
    FrontierServiceQueries.deleteOrganizationInvitation,
    {
      onSuccess: () => {
        handleOpenChange(false);
        toastManager.add({ title: 'Member deleted', type: 'success' });
      },
      onError: (error: Error) => {
        toastManager.add({
          title: 'Something went wrong',
          description: error?.message || 'Failed to delete invitation',
          type: 'error'
        });
      }
    }
  );

  const { mutateAsync: removeUser } = useMutation(
    FrontierServiceQueries.removeOrganizationUser,
    {
      onSuccess: () => {
        handleOpenChange(false);
        toastManager.add({ title: 'Member deleted', type: 'success' });
      },
      onError: (error: Error) => {
        toastManager.add({
          title: 'Something went wrong',
          description: error?.message || 'Failed to remove user',
          type: 'error'
        });
      }
    }
  );

  const deleteMember = async () => {
    setIsLoading(true);
    try {
      if (state.invited === 'true') {
        const req = create(DeleteOrganizationInvitationRequestSchema, {
          orgId: organizationId,
          id: state.memberId
        });
        await deleteInvitation(req);
      } else {
        const req = create(RemoveOrganizationUserRequestSchema, {
          id: organizationId,
          userId: state.memberId
        });
        await removeUser(req);
      }
    } catch (error: unknown) {
      toastManager.add({
        title: 'Something went wrong',
        description:
          error instanceof Error
            ? error.message
            : 'Failed to remove member',
        type: 'error'
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={state.open} onOpenChange={handleOpenChange}>
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
              onClick={() => handleOpenChange(false)}
              data-test-id="cancel-remove-member-dialog"
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="danger"
              onClick={deleteMember}
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
    </Dialog>
  );
});
