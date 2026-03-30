'use client';

import { forwardRef, useImperativeHandle, useState } from 'react';
import { create } from '@bufbuild/protobuf';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeletePolicyRequestSchema,
  CreatePolicyRequestSchema,
  ListPoliciesRequestSchema
} from '@raystack/proton/frontier';
import type { Role, Policy } from '@raystack/proton/frontier';
import {
  Button,
  Text,
  Dialog,
  Flex,
  toastManager
} from '@raystack/apsara-v1';

export interface UpdateRoleDialogHandle {
  open: (memberId: string, role: Role) => void;
}

export interface UpdateRoleDialogProps {
  organizationId: string;
  refetch: () => void;
}

export const UpdateRoleDialog = forwardRef<
  UpdateRoleDialogHandle,
  UpdateRoleDialogProps
>(function UpdateRoleDialog({ organizationId, refetch }, ref) {
  const [state, setState] = useState<{
    open: boolean;
    memberId: string;
    role: Role | null;
  }>({ open: false, memberId: '', role: null });
  const [isLoading, setIsLoading] = useState(false);

  useImperativeHandle(ref, () => ({
    open: (memberId: string, role: Role) =>
      setState({ open: true, memberId, role })
  }));

  const handleOpenChange = (value: boolean) => {
    if (!value) {
      setState({ open: false, memberId: '', role: null });
      refetch();
    }
  };

  const { data: policiesData } = useQuery(
    FrontierServiceQueries.listPolicies,
    create(ListPoliciesRequestSchema, {
      orgId: organizationId,
      userId: state.memberId
    }),
    { enabled: state.open && !!state.memberId && !!state.role }
  );

  const { mutateAsync: deletePolicy } = useMutation(
    FrontierServiceQueries.deletePolicy
  );

  const { mutateAsync: createPolicy } = useMutation(
    FrontierServiceQueries.createPolicy
  );

  const handleUpdate = async () => {
    if (!state.role) return;
    setIsLoading(true);
    try {
      const resource = `app/organization:${organizationId}`;
      const principal = `app/user:${state.memberId}`;
      const policies = policiesData?.policies || [];

      const deleteResults = await Promise.allSettled(
        policies.map((p: Policy) => {
          const req = create(DeletePolicyRequestSchema, {
            id: p.id as string
          });
          return deletePolicy(req);
        })
      );

      const deleteErrors = deleteResults
        .filter(
          (result): result is PromiseRejectedResult =>
            result.status === 'rejected'
        )
        .map(result => result.reason);

      if (deleteErrors.length > 0) {
        console.warn('Some policy deletions failed:', deleteErrors);
      }

      const createReq = create(CreatePolicyRequestSchema, {
        body: {
          roleId: state.role.id as string,
          title: state.role.name as string,
          resource,
          principal
        }
      });
      await createPolicy(createReq);

      toastManager.add({ title: 'Member role updated', type: 'success' });
      refetch();
      handleOpenChange(false);
    } catch (error: unknown) {
      toastManager.add({
        title: 'Something went wrong',
        description:
          error instanceof Error
            ? error.message
            : 'Failed to update member role',
        type: 'error'
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={state.open} onOpenChange={handleOpenChange}>
      <Dialog.Content width={400} showCloseButton={false}>
        <Dialog.Body>
          <Flex direction="column" gap={3}>
            <Text size="large" weight="medium">
              Update role
            </Text>
            <Text size="small" variant="secondary">
              This action will remove access to all projects where the user
              doesn&apos;t have an explicit project-level role.
            </Text>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              onClick={() => handleOpenChange(false)}
              disabled={isLoading}
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
            >
              Update
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
});
