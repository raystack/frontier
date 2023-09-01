import {
  Avatar,
  Button,
  Command,
  DataTable,
  DropdownMenu,
  EmptyState,
  Flex,
  Text
} from '@raystack/apsara';
import { useParams } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1User } from '~/src';
import { filterUsersfromUsers, getInitials } from '~/utils';
import { getColumns } from './member.columns';

export type MembersProps = {
  orgMembers?: V1Beta1User[];
  members: V1Beta1User[];
  setMembers: React.Dispatch<React.SetStateAction<V1Beta1User[]>>;
  organizationId?: string;
};

export const Members = ({
  orgMembers,
  members,
  setMembers,
  organizationId
}: MembersProps) => {
  const tableStyle = members?.length
    ? { width: '100%' }
    : { width: '100%', height: '100%' };

  const invitableUser = filterUsersfromUsers(orgMembers, members) || [];
  return (
    <Flex direction="column" style={{ paddingTop: '32px' }}>
      <DataTable
        data={members ?? []}
        // @ts-ignore
        columns={getColumns(organizationId)}
        emptyState={noDataChildren}
        parentStyle={{ height: 'calc(100vh - 212px)' }}
        style={tableStyle}
      >
        <DataTable.Toolbar style={{ padding: 0, border: 0 }}>
          <Flex justify="between" gap="small">
            <Flex style={{ maxWidth: '360px', width: '100%' }}>
              <DataTable.GloabalSearch
                placeholder="Search by name or email"
                size="medium"
              />
            </Flex>
            <DropdownMenu>
              <DropdownMenu.Trigger asChild>
                <Button variant="primary" style={{ width: 'fit-content' }}>
                  Add a member
                </Button>
              </DropdownMenu.Trigger>
              <DropdownMenu.Content
                align="end"
                style={{ minWidth: '220px', padding: 0 }}
              >
                <Command>
                  <Command.Input
                    placeholder="Add team member"
                    style={{ padding: '8px 0' }}
                  />
                  <Command.List>
                    <Command.Empty>No results found.</Command.Empty>
                    <Command.Group>
                      {invitableUser.length === 0 ? (
                        <Flex align="center" justify="center">
                          <Text>No member to invite</Text>
                        </Flex>
                      ) : null}
                      {invitableUser.map(user => (
                        <InviteUser
                          user={user}
                          key={user.id}
                          members={members}
                          setMembers={setMembers}
                          organizationId={organizationId}
                        />
                      ))}
                    </Command.Group>
                  </Command.List>
                </Command>
              </DropdownMenu.Content>
            </DropdownMenu>
          </Flex>
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
};

const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 members in your team</h3>
    <div className="pera">Try adding new members.</div>
  </EmptyState>
);

const InviteUser = ({
  user,
  organizationId,
  members = [],
  setMembers
}: {
  user: V1Beta1User;
  organizationId?: string;
  members?: V1Beta1User[];
  setMembers: React.Dispatch<React.SetStateAction<V1Beta1User[]>>;
}) => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const { client } = useFrontier();

  async function inviteMember() {
    try {
      await client?.frontierServiceAddGroupUsers(
        organizationId as string,
        teamId as string,
        { userIds: [user?.id as string] }
      );
      setMembers([...members, user]);
      toast.success('Member deleted');
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    }
  }

  return (
    <Flex
      align="center"
      gap="small"
      style={{ padding: 'var(--pd-8)', cursor: 'pointer' }}
      onClick={inviteMember}
    >
      <Avatar
        alt="profile"
        shape="square"
        fallback={getInitials(user?.title)}
        imageProps={{ width: '16px', height: '16px', fontSize: '8px' }}
      />
      <Text>{user.title}</Text>
    </Flex>
  );
};
