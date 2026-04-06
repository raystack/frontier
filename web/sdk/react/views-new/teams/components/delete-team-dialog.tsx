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
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteGroupRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

export interface DeleteTeamPayload {
  teamId: string;
}

type AlertDialogHandle = ReturnType<typeof AlertDialog.createHandle<DeleteTeamPayload>>;

export interface DeleteTeamDialogProps {
  handle: AlertDialogHandle;
  refetch: () => void;
}

export function DeleteTeamDialog({ handle, refetch }: DeleteTeamDialogProps) {
  const { activeOrganization: organization } = useFrontier();
  const [isLoading, setIsLoading] = useState(false);

  const { mutateAsync: deleteTeam } = useMutation(
    FrontierServiceQueries.deleteGroup
  );

  const handleDelete = async (teamId: string) => {
    if (!organization?.id || !teamId) return;
    setIsLoading(true);
    try {
      await deleteTeam(
        create(DeleteGroupRequestSchema, {
          id: teamId,
          orgId: organization.id
        })
      );
      toastManager.add({ title: 'Team deleted', type: 'success' });
      refetch();
      handle.close();
    } catch (error) {
      toastManager.add({
        title: 'Something went wrong',
        description:
          error instanceof Error ? error.message : 'Failed to delete team',
        type: 'error'
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AlertDialog handle={handle}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as DeleteTeamPayload | undefined;
        return (
          <AlertDialog.Content width={400}>
            <AlertDialog.Header>
              <AlertDialog.Title>Delete Team</AlertDialog.Title>
            </AlertDialog.Header>
            <AlertDialog.Body>
              <Text size="small" variant="secondary">
                This action is irreversible. All team data including member
                assignments will be permanently deleted. Are you sure you want
                to proceed?
              </Text>
            </AlertDialog.Body>
            <AlertDialog.Footer>
              <Flex gap={5} justify="end">
                <Button
                  variant="outline"
                  color="neutral"
                  onClick={() => handle.close()}
                  disabled={isLoading}
                  data-test-id="frontier-sdk-cancel-delete-team-btn"
                >
                  Cancel
                </Button>
                <Button
                  variant="solid"
                  color="danger"
                  onClick={() =>
                    payload && handleDelete(payload.teamId)
                  }
                  disabled={isLoading}
                  loading={isLoading}
                  loaderText="Deleting..."
                  data-test-id="frontier-sdk-delete-team-btn"
                >
                  Delete Now
                </Button>
              </Flex>
            </AlertDialog.Footer>
          </AlertDialog.Content>
        );
      }}
    </AlertDialog>
  );
}
