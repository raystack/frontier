import { Flex, Image, Text } from '@raystack/apsara';
import { Tabs } from '@raystack/apsara/v1';
import {
  Outlet,
  useNavigate,
  useParams,
  useRouterState
} from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { toast } from '@raystack/apsara/v1';
import backIcon from '~/react/assets/chevron-left.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1Role, V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { PERMISSIONS } from '~/utils';
import { General } from './general';
import { Members } from './members';
import { styles } from '../styles';
import orgStyles from '../organization.module.css';

export const TeamPage = () => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const [team, setTeam] = useState<V1Beta1Group>();
  const [members, setMembers] = useState<V1Beta1User[]>([]);
  const [memberRoles, setMemberRoles] = useState<Record<string, Role[]>>({});
  const [isTeamLoading, setIsTeamLoading] = useState(false);
  const [isMembersLoading, setIsMembersLoading] = useState(false);
  const [isTeamRoleLoading, setIsTeamRoleLoading] = useState(false);
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);

  const { client, activeOrganization: organization } = useFrontier();
  let navigate = useNavigate({ from: '/teams/$teamId' });
  const routerState = useRouterState();

  const isDeleteRoute = useMemo(() => {
    return routerState.matches.some(
      route => route.routeId === '/teams/$teamId/delete'
    );
  }, [routerState.matches]);

  useEffect(() => {
    async function getTeamDetails() {
      if (!organization?.id || !teamId || isDeleteRoute) return;

      try {
        setIsTeamLoading(true);
        const {
          // @ts-ignore
          data: { group }
        } = await client?.frontierServiceGetGroup(organization?.id, teamId);

        setTeam(group);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      } finally {
        setIsTeamLoading(false);
      }
    }
    getTeamDetails();
  }, [client, organization?.id, teamId, isDeleteRoute]);

  const getTeamMembers = useCallback(async () => {
    if (!organization?.id || !teamId || isDeleteRoute) return;
    try {
      setIsMembersLoading(true);
      const {
        // @ts-ignore
        data: { users, role_pairs }
      } = await client?.frontierServiceListGroupUsers(
        organization?.id,
        teamId,
        { withRoles: true }
      );
      setMembers(users);
      setMemberRoles(
        role_pairs.reduce((previous: any, mr: any) => {
          return { ...previous, [mr.user_id]: mr.roles };
        }, {})
      );
    } catch ({ error }: any) {
      toast.error('Something went wrong', {
        description: error.message
      });
    } finally {
      setIsMembersLoading(false);
    }
  }, [client, isDeleteRoute, organization?.id, teamId]);

  const getTeamRoles = useCallback(async () => {
    if (!organization?.id || !teamId || isDeleteRoute) return;
    try {
      setIsTeamRoleLoading(true);
      const {
        // @ts-ignore
        data: { roles }
      } = await client?.frontierServiceListRoles({
        state: 'enabled',
        scopes: [PERMISSIONS.GroupNamespace]
      });
      setRoles(roles);
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message
      });
    } finally {
      setIsTeamRoleLoading(false);
    }
  }, [client, isDeleteRoute, organization?.id, teamId]);

  useEffect(() => {
    getTeamMembers();
    getTeamRoles();
  }, [getTeamMembers, getTeamRoles]);

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
      <Tabs.Root defaultValue="general" className={orgStyles.orgTabsContainer} style={styles.container}>
        <Tabs.List>
          <Tabs.Trigger value="general">
            General
          </Tabs.Trigger>
          <Tabs.Trigger value="members">
            Members
          </Tabs.Trigger>
        </Tabs.List>
        <Tabs.Content value="general">
          <General
            organization={organization}
            team={team}
            isLoading={isTeamLoading}
          />
        </Tabs.Content>
        <Tabs.Content value="members">
          <Members
            members={members}
            roles={roles}
            memberRoles={memberRoles}
            organizationId={organization?.id || ''}
            isLoading={isMembersLoading}
            refetchMembers={getTeamMembers}
          />
        </Tabs.Content>
      </Tabs.Root>
      <Outlet />
    </Flex>
  );
};
