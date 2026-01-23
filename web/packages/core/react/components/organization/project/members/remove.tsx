import { Button, Flex, Text, toast, Image, Dialog } from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useMemo } from 'react';
import { useQuery, useMutation, createConnectQueryKey, useTransport } from '@connectrpc/connect-query';
import { useQueryClient } from '@tanstack/react-query';
import { FrontierServiceQueries, ListPoliciesRequestSchema, DeletePolicyRequestSchema, ListProjectUsersRequestSchema, ListProjectGroupsRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import cross from '~/react/assets/cross.svg';
import styles from '../../organization.module.css';

export const RemoveProjectMember = () => {
  const navigate = useNavigate({
    from: '/projects/$projectId/$membertype/$memberId/remove'
  });
  const queryClient = useQueryClient();
  const transport = useTransport();

  const { projectId, memberId } = useParams({
    from: '/projects/$projectId/$membertype/$memberId/remove'
  });

  const { data: policiesData, error: policiesError } = useQuery(
    FrontierServiceQueries.listPolicies,
    create(ListPoliciesRequestSchema, {
      projectId: projectId || '',
      userId: memberId || ''
    }),
    { enabled: !!projectId && !!memberId }
  );

  const policies = useMemo(() => policiesData?.policies ?? [], [policiesData]);

  useEffect(() => {
    if (policiesError) {
      toast.error('Something went wrong', { description: (policiesError as Error).message });
    }
  }, [policiesError]);

  const { mutateAsync: deletePolicy, isPending } = useMutation(FrontierServiceQueries.deletePolicy, {
    onError: (err: Error) =>
      toast.error('Something went wrong', { description: err.message })
  });

  async function onConfirm() {
    try {
      await Promise.all(
        (policies || []).map(p => deletePolicy(create(DeletePolicyRequestSchema, { id: p.id || '' })))
      );
      // Invalidate and refetch project users and groups queries after all policies are deleted
      if (projectId) {
        await queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.listProjectUsers,
            transport,
            input: create(ListProjectUsersRequestSchema, {
              id: projectId,
              withRoles: true
            }),
            cardinality: 'finite'
          })
        });
        await queryClient.invalidateQueries({
          queryKey: createConnectQueryKey({
            schema: FrontierServiceQueries.listProjectGroups,
            transport,
            input: create(ListProjectGroupsRequestSchema, {
              id: projectId,
              withRoles: true
            }),
            cardinality: 'finite'
          })
        });
      }
      navigate({ to: '/projects/$projectId', params: { projectId } });
      toast.success('Member removed');
    } catch (error) {
      // Error is already handled by mutation's onError callback
      // This catch prevents unhandled promise rejection
      console.error('Failed to delete policies:', error);
    }
  }

  return (
    <Dialog open={true}>
      <Dialog.Content
        style={{ padding: 0, maxWidth: '400px', width: '100%' }}
        overlayClassName={styles.overlay}
      >
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Remove project member
            </Text>
            <Image
              data-test-id="frontier-sdk-remove-project-member-close-btn"
              alt="cross"
              src={cross as unknown as string}
              onClick={() =>
                navigate({
                  to: '/projects/$projectId',
                  params: { projectId }
                })
              }
              style={{ cursor: 'pointer' }}
            />
          </Flex>
        </Dialog.Header>

        <Dialog.Body>
          <Flex direction="column" gap={5}>
            <Text size="regular">
              Are you sure you want to remove this member from the project?
            </Text>
          </Flex>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              size="normal"
              color="neutral"
              variant="outline"
              onClick={() => navigate({ to: '/members' })}
              data-test-id="frontier-sdk-remove-project-member-cancel-btn"
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button
              size="normal"
              color="danger"
              variant="solid"
              onClick={onConfirm}
              data-test-id="frontier-sdk-remove-project-member-confirm-btn"
              disabled={isPending}
              loading={isPending}
              loaderText="Removing"
            >
              Remove
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
