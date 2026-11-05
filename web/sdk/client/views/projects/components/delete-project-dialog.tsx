'use client';

import { useState } from 'react';
import {
  Button,
  Text,
  AlertDialog
} from '@raystack/apsara';
import { toastManager } from '@raystack/apsara';
import { useFrontier } from '../../../contexts/FrontierContext';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteProjectRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

export interface DeleteProjectPayload {
  projectId: string;
  projectName: string;
}

type AlertDialogHandle = ReturnType<typeof AlertDialog.createHandle<DeleteProjectPayload>>;

export interface DeleteProjectDialogProps {
  handle: AlertDialogHandle;
  refetch: () => void;
  onDeleteError: (error: unknown) => void;
}

export function DeleteProjectDialog({
  handle,
  refetch,
  onDeleteError
}: DeleteProjectDialogProps) {
  const { activeOrganization: organization } = useFrontier();
  const [isLoading, setIsLoading] = useState(false);

  const { mutateAsync: deleteProject } = useMutation(
    FrontierServiceQueries.deleteProject
  );

  const handleDelete = async (projectId: string) => {
    if (!organization?.id || !projectId) return;
    setIsLoading(true);
    try {
      await deleteProject(
        create(DeleteProjectRequestSchema, { id: projectId })
      );
      toastManager.add({ title: 'Project deleted', type: 'success' });
      refetch();
      handle.close();
    } catch (error) {
      onDeleteError(error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AlertDialog handle={handle}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as DeleteProjectPayload | undefined;
        return (
          <AlertDialog.Content>
            <AlertDialog.Header>
              <AlertDialog.Title>Delete Project</AlertDialog.Title>
            </AlertDialog.Header>
            <AlertDialog.Body>
              <Text size="small" variant="secondary">
                This action is irreversible and will delete all files
                associated with &quot;{payload?.projectName}.&quot; Are you
                sure you want to proceed?
              </Text>
            </AlertDialog.Body>
            <AlertDialog.Footer>
              <Button
                variant="outline"
                color="neutral"
                onClick={() => handle.close()}
                disabled={isLoading}
                data-test-id="frontier-sdk-cancel-delete-project-btn"
              >
                Cancel
              </Button>
              <Button
                variant="solid"
                color="danger"
                onClick={() =>
                  payload && handleDelete(payload.projectId)
                }
                disabled={isLoading}
                loading={isLoading}
                loaderText="Deleting..."
                data-test-id="frontier-sdk-delete-project-btn"
              >
                Delete Now
              </Button>
            </AlertDialog.Footer>
          </AlertDialog.Content>
        );
      }}
    </AlertDialog>
  );
}
