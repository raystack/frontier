'use client';

import { useState } from 'react';
import {
  Button,
  Flex,
  Text,
  AlertDialog
} from '@raystack/apsara-v1';
import { toastManager } from '@raystack/apsara-v1';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  RemoveProjectMemberRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { handleConnectError } from '~/utils/error';

export interface RemoveMemberPayload {
  memberId: string;
  projectId: string;
  memberType?: 'user' | 'group';
}

type AlertDialogHandle = ReturnType<
  typeof AlertDialog.createHandle<RemoveMemberPayload>
>;

export interface RemoveMemberDialogProps {
  handle: AlertDialogHandle;
  refetch: () => void;
}

export function RemoveMemberDialog({
  handle,
  refetch
}: RemoveMemberDialogProps) {
  return (
    <AlertDialog handle={handle}>
      {({ payload }) => {
        const p = payload as RemoveMemberPayload | undefined;
        return (
          <AlertDialog.Content width={400}>
            {p ? (
              <RemoveMemberForm
                payload={p}
                handle={handle}
                refetch={refetch}
              />
            ) : null}
          </AlertDialog.Content>
        );
      }}
    </AlertDialog>
  );
}

interface RemoveMemberFormProps {
  payload: RemoveMemberPayload;
  handle: AlertDialogHandle;
  refetch: () => void;
}

function RemoveMemberForm({
  payload,
  handle,
  refetch
}: RemoveMemberFormProps) {
  const [isLoading, setIsLoading] = useState(false);

  const { mutateAsync: removeProjectMember } = useMutation(
    FrontierServiceQueries.removeProjectMember
  );

  async function handleRemove() {
    setIsLoading(true);
    try {
      await removeProjectMember(
        create(RemoveProjectMemberRequestSchema, {
          projectId: payload.projectId,
          principalId: payload.memberId,
          principalType: payload.memberType === 'group' ? 'app/group' : 'app/user'
        })
      );
      toastManager.add({ title: 'Member removed', type: 'success' });
      refetch();
      handle.close();
    } catch (error) {
      handleConnectError(error, {
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: 'error' }),
        NotFound: (err) => toastManager.add({ title: 'Not found', description: err.message, type: 'error' }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.message, type: 'error' }),
      });
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <>
      <AlertDialog.Header>
        <AlertDialog.Title>Remove project member</AlertDialog.Title>
      </AlertDialog.Header>
      <AlertDialog.Body>
        <Text size="small" variant="secondary">
          Are you sure you want to remove this member from the project? This
          action cannot be undone.
        </Text>
      </AlertDialog.Body>
      <AlertDialog.Footer>
        <Flex gap={5} justify="end">
          <Button
            variant="outline"
            color="neutral"
            onClick={() => handle.close()}
            disabled={isLoading}
            data-test-id="frontier-sdk-cancel-remove-member-btn"
          >
            Cancel
          </Button>
          <Button
            variant="solid"
            color="danger"
            onClick={handleRemove}
            disabled={isLoading}
            loading={isLoading}
            loaderText="Removing..."
            data-test-id="frontier-sdk-remove-member-btn"
          >
            Remove
          </Button>
        </Flex>
      </AlertDialog.Footer>
    </>
  );
}
