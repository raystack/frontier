import { useEffect, useMemo } from 'react';
import { Tabs, Image, Text, Flex, toast } from '@raystack/apsara';
import {
  Outlet,
  useNavigate,
  useParams,
  useRouterState
} from '@tanstack/react-router';
import backIcon from '~/react/assets/chevron-left.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { PERMISSIONS } from '~/utils';
import { General } from './general';
import { Members } from './members';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, GetGroupRequestSchema, ListGroupUsersRequestSchema, ListRolesRequestSchema, Organization } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import styles from './teams.module.css';

export const TeamPage = () => {
  let { teamId } = useParams({ from: '/teams/$teamId' });

  const { activeOrganization: organization } = useFrontier();
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
    create(GetGroupRequestSchema, { id: teamId || '', orgId: organization?.id || '', withMembers: true }),
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

  // Get team members using Connect RPC
  const { data: membersData, isLoading: isMembersLoading, error: membersError, refetch: refetchMembers } = useQuery(
    FrontierServiceQueries.listGroupUsers,
    create(ListGroupUsersRequestSchema, { id: teamId || '', orgId: organization?.id || '', withRoles: true }),
    { enabled: !!organization?.id && !!teamId && !isDeleteRoute }
  );

  const members = membersData?.users || [];
  const memberRoles = useMemo(() => {
    if (!membersData?.rolePairs) return {};
    return membersData.rolePairs.reduce((previous: any, mr: any) => {
      return { ...previous, [mr.user_id]: mr.roles };
    }, {});
  }, [membersData?.rolePairs]);

  // Handle members error
  useEffect(() => {
    if (membersError) {
      toast.error('Something went wrong', {
        description: membersError.message
      });
    }
  }, [membersError]);

  // Get team roles using Connect RPC
  const { data: rolesData, error: rolesError } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, { state: 'enabled', scopes: [PERMISSIONS.GroupNamespace] }),
    { enabled: !!organization?.id && !!teamId && !isDeleteRoute }
  );

  const roles = rolesData?.roles || [];

  // Handle roles error
  useEffect(() => {
    if (rolesError) {
      toast.error('Something went wrong', {
        description: rolesError.message
      });
    }
  }, [rolesError]);

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
            organization={organization as Organization}
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
            refetchMembers={refetchMembers}
          />
        </Tabs.Content>
      </Tabs>
      <Outlet />
    </Flex>
  );
};
