import { Button, toast, Image, Text, Dialog, Flex } from '@raystack/apsara';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useState } from 'react';
import { useTerminology } from '~/react/hooks/useTerminology';
import { useMutation } from '@connectrpc/connect-query';
import {
  FrontierServiceQueries,
  DeleteOrganizationInvitationRequestSchema,
  RemoveOrganizationMemberRequestSchema
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';

export interface MemberRemoveConfirmDialogProps {
  open: boolean;
  onOpenChange: (value: boolean) => void;
  memberId: string;
  invited: string;
}

export const RemoveMemberDialog = ({
  open,
  onOpenChange,
  memberId,
  invited
}: MemberRemoveConfirmDialogProps) => {
  const { activeOrganization } = useFrontier();
  const organizationId = activeOrganization?.id ?? '';
  const [isLoading, setIsLoading] = useState(false);
  const t = useTerminology();

  const { mutateAsync: deleteInvitation } = useMutation(
    FrontierServiceQueries.deleteOrganizationInvitation,
    {
      onSuccess: () => {
        onOpenChange(false);
        toast.success('Invitation deleted');
      },
      onError: (error: any) => {
        toast.error('Something went wrong', {
          description: error?.message || 'Failed to delete invitation'
        });
      }
    }
  );

  const { mutateAsync: removeMember } = useMutation(
    FrontierServiceQueries.removeOrganizationMember,
    {
      onSuccess: () => {
        onOpenChange(false);
        toast.success('User removed');
      },
      onError: (error: any) => {
        toast.error('Something went wrong', {
          description: error?.message || 'Failed to remove user'
        });
      }
    }
  );

  const deleteMember = async () => {
    setIsLoading(true);
    try {
      if (invited === 'true') {
        const req = create(DeleteOrganizationInvitationRequestSchema, {
          orgId: organizationId,
          id: memberId
        });
        await deleteInvitation(req);
      } else {
        const req = create(RemoveOrganizationMemberRequestSchema, {
          orgId: organizationId,
          principalId: memberId,
          principalType: 'app/user'
        });
        await removeMember(req);
      }
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message || 'Failed to remove member'
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Content style={{ padding: 0, maxWidth: '400px', width: '100%' }}>
        <Dialog.Header>
          <Flex justify="between" align="center" style={{ width: '100%' }}>
            <Text size="large" weight="medium">
              Remove member?
            </Text>
            <Image
              alt="cross"
              src={cross as unknown as string}
              onClick={() => (isLoading ? null : onOpenChange(false))}
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
              onClick={() => onOpenChange(false)}
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
