import { Flex, Text, Image } from '@raystack/apsara';

import { Tabs } from '@raystack/apsara';
import {
  Outlet,
  useNavigate,
  useParams,
  useRouterState
} from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1User } from '~/src';
import { styles } from '../styles';
import { General } from './general';
import { Members } from './members';
import backIcon from '~/react/assets/chevron-left.svg';

export const TeamPage = () => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const [team, setTeam] = useState<V1Beta1Group>();
  const [members, setMembers] = useState<V1Beta1User[]>([]);
  const { client, activeOrganization: organization } = useFrontier();
  let navigate = useNavigate({ from: '/teams/$teamId' });
  const routerState = useRouterState();

  useEffect(() => {
    async function getTeamDetails() {
      if (!organization?.id || !teamId) return;

      try {
        const {
          // @ts-ignore
          data: { group }
        } = await client?.frontierServiceGetGroup(organization?.id, teamId);

        setTeam(group);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    }
    getTeamDetails();
  }, [client, organization?.id, teamId]);

  useEffect(() => {
    async function getTeamMembers() {
      if (!organization?.id || !teamId) return;
      try {
        const {
          // @ts-ignore
          data: { users, role_pairs }
        } = await client?.frontierServiceListGroupUsers(
          organization?.id,
          teamId,
          { withRoles: true }
        );
        setMembers(users);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    }
    getTeamMembers();
  }, [client, organization?.id, teamId, routerState.location.key]);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Image
          alt="back-icon"
          style={{ cursor: 'pointer' }}
          // @ts-ignore
          src={backIcon}
          onClick={() => navigate({ to: '/teams' })}
        />
        <Text size={6}>Teams</Text>
      </Flex>
      <Tabs defaultValue="general" style={styles.container}>
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
          <Members members={members} organizationId={organization?.id} />
        </Tabs.Content>
      </Tabs>
      <Outlet />
    </Flex>
  );
};
