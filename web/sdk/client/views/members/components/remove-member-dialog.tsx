'use client';

import { useState } from 'react';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteOrganizationInvitationRequestSchema,
  RemoveOrganizationMemberRequestSchema
} from '@raystack/proton/frontier';
import {
  AlertDialog,
  Button,
  toastManager
} from '@raystack/apsara';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useTerminology } from '../../../hooks/useTerminology';
import { handleConnectError } from '~/utils/error';

export type RemoveMemberPayload = { memberId: string; invited: string };

export interface RemoveMemberDialogProps {
  handle: ReturnType<typeof AlertDialog.createHandle<RemoveMemberPayload>>;
  refetch: () => void;
}

export function RemoveMemberDialog({ handle, refetch }: RemoveMemberDialogProps) {
  const { activeOrganization } = useFrontier();
  const organizationId = activeOrganization?.id ?? '';
  const [isLoading, setIsLoading] = useState(false);
  const t = useTerminology();

  const { mutateAsync: deleteInvitation } = useMutation(
    FrontierServiceQueries.deleteOrganizationInvitation,
    {
      onSuccess: () => {
        handle.close();
        refetch();
        toastManager.add({ title: 'Invitation deleted', type: 'success' });
      }
    }
  );

  const { mutateAsync: removeMember } = useMutation(
    FrontierServiceQueries.removeOrganizationMember,
    {
      onSuccess: () => {
        handle.close();
        refetch();
        toastManager.add({ title: 'User removed', type: 'success' });
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
        const req = create(RemoveOrganizationMemberRequestSchema, {
          orgId: organizationId,
          principalId: memberId,
          principalType: 'app/user'
        });
        await removeMember(req);
      }
      handle.close();
    } catch (error) {
      handleConnectError(error, {
        NotFound: (err) => toastManager.add({ title: 'Not found', description: err.message, type: 'error' }),
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
        FailedPrecondition: (err) => toastManager.add({ title: 'Cannot remove user', description: err.message, type: 'error' }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AlertDialog handle={handle}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as RemoveMemberPayload | undefined;
        return (
          <AlertDialog.Content>
            <AlertDialog.Header>
              <AlertDialog.Title>Remove member</AlertDialog.Title>
              <AlertDialog.Description>
                Are you sure you want to remove this member from the{' '}
                {t.organization({ case: 'lower' })}?
              </AlertDialog.Description>
            </AlertDialog.Header>
            <AlertDialog.Footer>
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
            </AlertDialog.Footer>
          </AlertDialog.Content>
        );
      }}
    </AlertDialog>
  );
}
