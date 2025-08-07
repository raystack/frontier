import { Button, Flex, Text, toast, Image, Dialog } from '@raystack/apsara';
import cross from '~/react/assets/cross.svg';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useState } from 'react';
import { V1Beta1Policy } from '~/api-client';
import styles from '../../organization.module.css';

export const RemoveProjectMember = () => {
  const [isLoading, setIsLoading] = useState(false);

  const { client } = useFrontier();
  const navigate = useNavigate({
    from: '/projects/$projectId/$membertype/$memberId/remove'
  });

  const { projectId, memberId } = useParams({
    from: '/projects/$projectId/$membertype/$memberId/remove'
  });

  async function onConfirm() {
    setIsLoading(true);
    try {
      const {
        // @ts-ignore
        data: { policies = [] }
      } = await client?.frontierServiceListPolicies({
        project_id: projectId,
        user_id: memberId
      });

      const deletePromises = policies.map((p: V1Beta1Policy) =>
        client?.frontierServiceDeletePolicy(p.id as string)
      );

      await Promise.all(deletePromises);
      navigate({
        to: '/projects/$projectId',
        params: { projectId },
        state: { refetch: true }
      });
      toast.success('Member removed');
    } catch (err: any) {
      toast.error('Something went wrong', {
        description: err.message
      });
    } finally {
      setIsLoading(false);
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
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              size="normal"
              color="danger"
              variant="solid"
              onClick={onConfirm}
              data-test-id="frontier-sdk-remove-project-member-confirm-btn"
              disabled={isLoading}
              loading={isLoading}
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
