'use client';

import { useState } from 'react';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  SetOrganizationMemberRoleRequestSchema
} from '@raystack/proton/frontier';
import type { Role } from '@raystack/proton/frontier';
import {
  AlertDialog,
  Button,
  toastManager
} from '@raystack/apsara-v1';
import { handleConnectError } from '~/utils/error';

export type UpdateRolePayload = { memberId: string; role: Role };

export interface UpdateRoleDialogProps {
  handle: ReturnType<typeof AlertDialog.createHandle<UpdateRolePayload>>;
  organizationId: string;
  refetch: () => void;
}

export function UpdateRoleDialog({ handle, organizationId, refetch }: UpdateRoleDialogProps) {
  return (
    <AlertDialog handle={handle}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as UpdateRolePayload | undefined;
        return payload ? (
          <UpdateRoleContent
            payload={payload}
            organizationId={organizationId}
            onClose={() => handle.close()}
            refetch={refetch}
          />
        ) : null;
      }}
    </AlertDialog>
  );
}

function UpdateRoleContent({
  payload,
  organizationId,
  onClose,
  refetch
}: {
  payload: UpdateRolePayload;
  organizationId: string;
  onClose: () => void;
  refetch: () => void;
}) {
  const [isLoading, setIsLoading] = useState(false);

  const { mutateAsync: setMemberRole } = useMutation(
    FrontierServiceQueries.setOrganizationMemberRole
  );

  const handleUpdate = async () => {
    setIsLoading(true);
    try {
      const req = create(SetOrganizationMemberRoleRequestSchema, {
        orgId: organizationId,
        userId: payload.memberId,
        roleId: payload.role.id as string
      });
      await setMemberRole(req);

      toastManager.add({ title: 'Member role updated', type: 'success' });
      refetch();
      onClose();
    } catch (error) {
      handleConnectError(error, {
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
        NotFound: (err) => toastManager.add({ title: 'Not found', description: err.message, type: 'error' }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AlertDialog.Content>
      <AlertDialog.Header>
        <AlertDialog.Title>Update role</AlertDialog.Title>
        <AlertDialog.Description>
          This will grant additional permissions to the user based on the new
          role.
        </AlertDialog.Description>
      </AlertDialog.Header>
      <AlertDialog.Footer>
        <Button
          variant="outline"
          color="neutral"
          onClick={onClose}
          disabled={isLoading}
          data-test-id="frontier-sdk-cancel-update-role-dialog"
        >
          Cancel
        </Button>
        <Button
          variant="solid"
          color="accent"
          onClick={handleUpdate}
          disabled={isLoading}
          loading={isLoading}
          loaderText="Updating..."
          data-test-id="frontier-sdk-confirm-update-role-dialog"
        >
          Update
        </Button>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  );
}
