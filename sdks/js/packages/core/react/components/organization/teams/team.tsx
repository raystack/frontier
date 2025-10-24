import { useCallback, useEffect, useMemo, useState } from 'react';
import { Tabs, Image, Text, Flex, toast } from '@raystack/apsara';
import {
  Outlet,
  useNavigate,
  useParams,
  useRouterState
} from '@tanstack/react-router';
import backIcon from '~/react/assets/chevron-left.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import type { V1Beta1Role, V1Beta1User } from '~/src';
import type { Role } from '~/src/types';
import { PERMISSIONS } from '~/utils';
import { General } from './general';
import { Members } from './members';
import styles from './teams.module.css';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries } from '@raystack/proton/frontier';

export const TeamPage = () => {
  let { teamId } = useParams({ from: '/teams/$teamId' });
  const [members, setMembers] = useState<V1Beta1User[]>([]);
  const [memberRoles, setMemberRoles] = useState<Record<string, Role[]>>({});
  const [isMembersLoading, setIsMembersLoading] = useState(false);
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);

  const { client, activeOrganization: organization } = useFrontier();
  let navigate = useNavigate({ from: '/teams/$teamId' });
  const routerState = useRouterState();

  const isDeleteRoute = useMemo(() => {
    return routerState.matches.some(
      route => route.routeId === '/teams/$teamId/delete'
    );
  }, [routerState.matches]);

  // Get team details using Connect RPC
  const { data: teamData, isLoading: isTeamLoading, error: teamError } = useQuery(
    FrontierServiceQueries.getGroup,
    { id: teamId || '', orgId: organization?.id || '', withMembers: true },
    { enabled: !!organization?.id && !!teamId && !isDeleteRoute }
  );

  const team = teamData?.group;

  // Handle team error
  useEffect(() => {
    if (teamError) {
      toast.error('Something went wrong', {
        description: teamError.message
      });
    }
  }, [teamError]);

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
    }
  }, [client, isDeleteRoute, organization?.id, teamId]);

  useEffect(() => {
    getTeamMembers();
    getTeamRoles();
  }, [getTeamMembers, getTeamRoles]);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex className={styles.header}>
        <Image
          alt="back-icon"
          style={{ cursor: 'pointer' }}
          src={backIcon as unknown as string}
          onClick={() => navigate({ to: '/teams' })}
          data-test-id="frontier-sdk-team-back-btn"
        />
        <Text size="large">Teams</Text>
      </Flex>
      <Tabs defaultValue="general" className={styles.container}>
        <Tabs.List>
          <Tabs.Trigger value="general">General</Tabs.Trigger>
          <Tabs.Trigger value="members">Members</Tabs.Trigger>
        </Tabs.List>
        <Tabs.Content value="general">
          <General
            organization={organization}
            team={team}
            isLoading={isTeamLoading}
          />
        </Tabs.Content>
        <Tabs.Content value="members" className={styles.tabContent}>
          <Members
            members={members}
            roles={roles}
            memberRoles={memberRoles}
            organizationId={organization?.id || ''}
            isLoading={isMembersLoading}
            refetchMembers={getTeamMembers}
          />
        </Tabs.Content>
      </Tabs>
      <Outlet />
    </Flex>
  );
};
