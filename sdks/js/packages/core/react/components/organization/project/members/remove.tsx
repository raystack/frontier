import { Button, Flex, Text, toast, Image, Dialog } from '@raystack/apsara';
import cross from '~/react/assets/cross.svg';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect } from 'react';
import { useQuery, useMutation } from '@connectrpc/connect-query';
import { FrontierServiceQueries, ListPoliciesRequestSchema, DeletePolicyRequestSchema } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import styles from '../../organization.module.css';

export const RemoveProjectMember = () => {

  const navigate = useNavigate({
    from: '/projects/$projectId/$membertype/$memberId/remove'
  });

  const { projectId, memberId } = useParams({
    from: '/projects/$projectId/$membertype/$memberId/remove'
  });

  const { data: policies = [], error: policiesError } = useQuery(
    FrontierServiceQueries.listPolicies,
    create(ListPoliciesRequestSchema, {
      projectId: projectId || '',
      userId: memberId || ''
    }),
    { enabled: !!projectId && !!memberId, select: d => d?.policies ?? [] }
  );

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
    await Promise.all(
      (policies || []).map(p => deletePolicy(create(DeletePolicyRequestSchema, { id: p.id || '' })))
    );
    navigate({ to: '/projects/$projectId', params: { projectId }, state: { refetch: true } });
    toast.success('Member removed');
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
