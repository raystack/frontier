import {
    Flex,
    Text,
    Dialog,
    Button,
    Separator,
    Image
  } from '@raystack/apsara';
  import cross from '~/react/assets/cross.svg';

const MemberRemoveConfirm = ({ isOpen, setIsOpen, deleteMember, isLoading }: {
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
  deleteMember: () => void;
  isLoading: boolean;
}) => {
  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen} >
      <Dialog.Content style={{ padding: 0, maxWidth: '400px', width: '100%', zIndex: '60' }}>
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} style={{ fontWeight: '500' }}>
            Remove member?
          </Text>
          <Image
            alt="cross"
            src={cross}
            onClick={() => isLoading ? null : setIsOpen(false)}
            style={{ cursor: isLoading ? 'not-allowed' : 'pointer' }}
            data-test-id="close-remove-member-dialog"
          />
        </Flex>
        <Separator />
        <Flex direction="column" gap="medium" style={{ padding: '24px' }}>
          <Text size={4}>
            Are you sure you want to remove this member from the organization?
          </Text>
        </Flex>
        <Separator />
        <Flex justify="end" style={{ padding: 'var(--pd-16)' }} gap="medium">
          <Button
            size="medium"
            variant="secondary"
            onClick={() => setIsOpen(false)}
            data-test-id="cancel-remove-member-dialog"
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            size="medium"
            variant="danger"
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