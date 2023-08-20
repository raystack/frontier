import { Flex, Text } from '@raystack/apsara';

import { Tabs } from '@raystack/apsara';
import { useEffect, useState } from 'react';
import { Outlet, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1Organization, V1Beta1User } from '~/src';
import { styles } from '../styles';
import { General } from './general';
import { Members } from './members';

interface TeamPageProps {
  organization?: V1Beta1Organization;
}
export const TeamPage = ({ organization }: TeamPageProps) => {
  let { teamId } = useParams();
  const [team, setTeam] = useState<V1Beta1Group>();
  const [members, setMembers] = useState<V1Beta1User[]>([]);
  const { client } = useFrontier();

  useEffect(() => {
    async function getTeamDetails() {
      if (!organization?.id || !teamId) return;

      try {
        const {
          // @ts-ignore
          data: { group }
        } = await client?.frontierServiceGetGroup(organization?.id, teamId);
        const {
          // @ts-ignore
          data: { users }
        } = await client?.frontierServiceListGroupUsers(
          organization?.id,
          teamId
        );
        setTeam(group);
        setMembers(users);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    }
    getTeamDetails();
  }, [client, organization?.id, teamId]);

  return (
    <Flex direction="column" gap="large" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Teams</Text>
      </Flex>
      <Tabs defaultValue="general" style={{ margin: '0 48px', zIndex: 0 }}>
        <Tabs.List elevated>
          <Tabs.Trigger value="general" style={{ flex: 1, height: 24 }}>
            General
          </Tabs.Trigger>
          <Tabs.Trigger value="members" style={{ flex: 1, height: 24 }}>
            Members
          </Tabs.Trigger>
        </Tabs.List>
        <Tabs.Content value="general">
          <General organization={organization} team={team} />
        </Tabs.Content>
        <Tabs.Content value="members">
          <Members members={members} />
        </Tabs.Content>
      </Tabs>
      <Outlet />
    </Flex>
  );
};
