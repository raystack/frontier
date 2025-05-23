import {
  Flex,
  Dialog,
} from '@raystack/apsara';
import { Button, Separator, toast, Image, Text } from '@raystack/apsara/v1';
import cross from '~/react/assets/cross.svg';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useState } from 'react';

const MemberRemoveConfirm = () => {
  const navigate = useNavigate({ from: '/members/remove-member/$memberId/$invited' });
  const { memberId, invited } = useParams({ from: '/members/remove-member/$memberId/$invited' });
  const { client, activeOrganization } = useFrontier();
  const organizationId = activeOrganization?.id ?? ''
  const [isLoading, setIsLoading] = useState(false);

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
      <Dialog.Content style={{ padding: 0, maxWidth: '400px', width: '100%', zIndex: '60' }}>
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size="large" weight="medium">
            Remove member?
          </Text>
          <Image
            alt="cross"
            src={cross as unknown as string}
            onClick={() => isLoading ? null : navigate({ to: '/members' })}
            style={{ cursor: isLoading ? 'not-allowed' : 'pointer' }}
            data-test-id="close-remove-member-dialog"
          />
        </Flex>
        <Separator />
        <Flex direction="column" gap="medium" style={{ padding: '24px' }}>
          <Text size="regular">
            Are you sure you want to remove this member from the organization?
          </Text>
        </Flex>
        <Separator />
        <Flex justify="end" style={{ padding: 'var(--pd-16)' }} gap="medium">
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
      </Dialog.Content>
    </Dialog>
  )
}

export default MemberRemoveConfirm