import { Button, toast, Image, Text, Dialog, Flex } from '@raystack/apsara';
import cross from '~/react/assets/cross.svg';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useState } from 'react';
import { useTerminology } from '~/react/hooks/useTerminology';

const MemberRemoveConfirm = () => {
  const navigate = useNavigate({
    from: '/members/remove-member/$memberId/$invited'
  });
  const { memberId, invited } = useParams({
    from: '/members/remove-member/$memberId/$invited'
  });
  const { client, activeOrganization } = useFrontier();
  const organizationId = activeOrganization?.id ?? '';
  const [isLoading, setIsLoading] = useState(false);
  const t = useTerminology();
  const deleteMember = async () => {
    setIsLoading(true);
    try {
      if (invited === 'true') {
        await client?.frontierServiceDeleteOrganizationInvitation(
          organizationId,
          memberId as string
        );
      } else {
        await client?.frontierServiceRemoveOrganizationUser(
          organizationId,
          memberId as string
        );
      }
      navigate({ to: '/members' });
      toast.success('Member deleted');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={true} onOpenChange={() => navigate({ to: '/members' })}>
      <Dialog.Content style={{ padding: 0, maxWidth: '400px', width: '100%' }}>
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Remove member?
            </Text>
            <Image
              alt="cross"
              src={cross as unknown as string}
              onClick={() => (isLoading ? null : navigate({ to: '/members' }))}
              style={{ cursor: isLoading ? 'not-allowed' : 'pointer' }}
              data-test-id="close-remove-member-dialog"
            />
          </Flex>
        </Dialog.Header>

        <Dialog.Body>
          <Flex direction="column" gap={5}>
            <Text size="regular">
              Are you sure you want to remove this member from the{' '}
              {t.organization({ case: 'lower' })}?
            </Text>
          </Flex>
        </Dialog.Body>

        <Dialog.Footer>
          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              onClick={() => navigate({ to: '/members' })}
              data-test-id="cancel-remove-member-dialog"
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="danger"
              onClick={deleteMember}
              data-test-id="confirm-remove-member-dialog"
              disabled={isLoading}
            >
              {isLoading ? 'Removing...' : 'Remove'}
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

export default MemberRemoveConfirm;
