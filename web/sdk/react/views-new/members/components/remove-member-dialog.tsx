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

export interface RemoveMemberDialogProps {
  open: boolean;
  onOpenChange: (value: boolean) => void;
  memberId: string;
  invited: string;
}

export function RemoveMemberDialog({
  open,
  onOpenChange,
  memberId,
  invited
}: RemoveMemberDialogProps) {
  const { activeOrganization } = useFrontier();
  const organizationId = activeOrganization?.id ?? '';
  const [isLoading, setIsLoading] = useState(false);
  const t = useTerminology();

  const { mutateAsync: deleteInvitation } = useMutation(
    FrontierServiceQueries.deleteOrganizationInvitation,
    {
      onSuccess: () => {
        onOpenChange(false);
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
        onOpenChange(false);
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
    <Dialog open={open} onOpenChange={onOpenChange}>
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
              onClick={() => onOpenChange(false)}
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
}
