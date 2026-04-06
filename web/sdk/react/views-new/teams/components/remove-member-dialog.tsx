'use client';

import { useState } from 'react';
import {
  Button,
  Flex,
  Text,
  AlertDialog
} from '@raystack/apsara-v1';
import { toastManager } from '@raystack/apsara-v1';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  RemoveGroupUserRequestSchema,
  ListPoliciesRequestSchema,
  DeletePolicyRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

export interface RemoveMemberPayload {
  memberId: string;
  teamId: string;
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
  const { activeOrganization: organization } = useFrontier();

  const { data: policiesData } = useQuery(
    FrontierServiceQueries.listPolicies,
    create(ListPoliciesRequestSchema, {
      groupId: payload.teamId,
      userId: payload.memberId
    }),
    { enabled: !!payload.teamId && !!payload.memberId }
  );

  const policies = policiesData?.policies ?? [];

  const { mutateAsync: deletePolicy } = useMutation(
    FrontierServiceQueries.deletePolicy
  );

  const { mutateAsync: removeGroupUser } = useMutation(
    FrontierServiceQueries.removeGroupUser
  );

  async function handleRemove() {
    if (!organization?.id) return;
    setIsLoading(true);
    try {
      await Promise.all(
        policies.map(p =>
          deletePolicy(create(DeletePolicyRequestSchema, { id: p.id || '' }))
        )
      );

      await removeGroupUser(
        create(RemoveGroupUserRequestSchema, {
          id: payload.teamId,
          orgId: organization.id,
          userId: payload.memberId
        })
      );

      toastManager.add({ title: 'Member removed', type: 'success' });
      refetch();
      handle.close();
    } catch (error) {
      toastManager.add({
        title: 'Something went wrong',
        description:
          error instanceof Error ? error.message : 'Failed to remove member',
        type: 'error'
      });
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <>
      <AlertDialog.Header>
        <AlertDialog.Title>Remove team member</AlertDialog.Title>
      </AlertDialog.Header>
      <AlertDialog.Body>
        <Text size="small" variant="secondary">
          Are you sure you want to remove this member from the team? This
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
            data-test-id="frontier-sdk-cancel-remove-team-member-btn"
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
            data-test-id="frontier-sdk-remove-team-member-btn"
          >
            Remove
          </Button>
        </Flex>
      </AlertDialog.Footer>
    </>
  );
}
