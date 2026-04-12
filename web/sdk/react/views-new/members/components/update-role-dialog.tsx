'use client';

import { useState } from 'react';
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
import { handleConnectError } from '~/utils/error';

export type UpdateRolePayload = { memberId: string; role: Role };

export interface UpdateRoleDialogProps {
  handle: ReturnType<typeof Dialog.createHandle<UpdateRolePayload>>;
  organizationId: string;
  refetch: () => void;
}

export function UpdateRoleDialog({ handle, organizationId, refetch }: UpdateRoleDialogProps) {
  const handleOpenChange = (open: boolean) => {
    if (!open) {
      refetch();
    }
  };

  return (
    <Dialog handle={handle} onOpenChange={handleOpenChange}>
      {({ payload: rawPayload }) => {
        const payload = rawPayload as UpdateRolePayload | undefined;
        return payload ? (
          <UpdateRoleContent
            payload={payload}
            organizationId={organizationId}
            onClose={() => handle.close()}
          />
        ) : null;
      }}
    </Dialog>
  );
}

function UpdateRoleContent({
  payload,
  organizationId,
  onClose
}: {
  payload: UpdateRolePayload;
  organizationId: string;
  onClose: () => void;
}) {
  const [isLoading, setIsLoading] = useState(false);

  const { data: policiesData } = useQuery(
    FrontierServiceQueries.listPolicies,
    create(ListPoliciesRequestSchema, {
      orgId: organizationId,
      userId: payload.memberId
    }),
    { enabled: !!payload.memberId && !!payload.role }
  );

  const { mutateAsync: deletePolicy } = useMutation(
    FrontierServiceQueries.deletePolicy
  );

  const { mutateAsync: createPolicy } = useMutation(
    FrontierServiceQueries.createPolicy
  );

  const handleUpdate = async () => {
    setIsLoading(true);
    try {
      const resource = `app/organization:${organizationId}`;
      const principal = `app/user:${payload.memberId}`;
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
          roleId: payload.role.id as string,
          title: payload.role.name as string,
          resource,
          principal
        }
      });
      await createPolicy(createReq);

      toastManager.add({ title: 'Member role updated', type: 'success' });
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
            onClick={onClose}
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
  );
}
