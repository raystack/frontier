import {
  Button,
  Dialog,
  Flex,
  Image,
  Separator,
  Text,
  TextField
} from '@raystack/apsara';
import { useNavigate } from 'react-router-dom';
import cross from '~/react/assets/cross.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Organization, V1Beta1User } from '~/src';

export const InviteMember = ({
  organization,
  users = []
}: {
  organization?: V1Beta1Organization;
  users?: V1Beta1User[];
}) => {
  const navigate = useNavigate();
  const { client } = useFrontier();

  return (
    <Dialog open={true}>
      <Dialog.Content style={{ padding: 0, maxWidth: '600px', width: '100%' }}>
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} style={{ fontWeight: '500' }}>
            Add people to the team
          </Text>
          {/* @ts-ignore */}
          <Image alt="cross" src={cross} onClick={() => navigate('/members')} />
        </Flex>
        <Separator />
        <Flex direction="column" gap="medium" style={{ padding: '24px 32px' }}>
          {/* @ts-ignore */}
          <TextField placeholder="Search organisation member" size="medium" />
          <Flex direction="column" gap="small">
            {users.map(u => (
              <InvitableUser key={u.id} user={u} />
            ))}
          </Flex>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
};

const InvitableUser = ({ user }: { user: V1Beta1User }) => {
  return (
    <Flex
      justify="between"
      align="center"
      style={{
        minHeight: '200px',
        maxHeight: '320px',
        height: '100%',
        overflow: 'scroll'
      }}
    >
      <Flex direction="column" gap="extra-small">
        <Text size={4} style={{ fontWeight: 500 }}>
          {user.name}
        </Text>
        <Text size={2}>{user.name}</Text>
      </Flex>
      <Button
        variant="secondary"
        size="small"
        style={{ height: 'fit-content', width: 'fit-content' }}
      >
        Add
      </Button>
    </Flex>
  );
};
