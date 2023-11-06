import { Flex, Image, Text } from '@raystack/apsara';

import { Tabs } from '@raystack/apsara';
import {
  Outlet,
  useNavigate,
  useParams,
  useRouterState
} from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';
import backIcon from '~/react/assets/chevron-left.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Group, V1Beta1Project, V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { styles } from '../styles';
import { General } from './general';
import { Members } from './members';

export const ProjectPage = () => {
  let { projectId } = useParams({ from: '/projects/$projectId' });
  const [isProjectLoading, setIsProjectLoading] = useState(false);
  const [project, setProject] = useState<V1Beta1Project>();
  const [members, setMembers] = useState<V1Beta1User[]>([]);
  const [memberRoles, setMemberRoles] = useState<Record<string, Role[]>>({});
  const [isMembersLoading, setIsMembersLoading] = useState(false);

  const [teams, setTeams] = useState<V1Beta1Group[]>([]);
  const [isTeamsLoading, setIsTeamsLoading] = useState(false);

  const { client, activeOrganization: organization } = useFrontier();
  let navigate = useNavigate({ from: '/projects/$projectId' });
  const routeState = useRouterState();

  const isDeleteRoute = useMemo(() => {
    return routeState.matches.some(
      route => route.routeId === '/projects/$projectId/delete'
    );
  }, [routeState.matches]);

  const getProjectTeams = useCallback(async () => {
    if (!organization?.id || !projectId || isDeleteRoute) return;
    try {
      setIsTeamsLoading(true);
      const result = await client?.frontierServiceListProjectGroups(projectId);
      if (result) {
        const {
          data: { groups = [] }
        } = result;
        setTeams(groups);
      }
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message
      });
    } finally {
      setIsTeamsLoading(false);
    }
  }, [client, isDeleteRoute, organization?.id, projectId]);

  const getProjectMembers = useCallback(async () => {
    if (!organization?.id || !projectId || isDeleteRoute) return;
    try {
      setIsMembersLoading(true);
      const {
        // @ts-ignore
        data: { users, role_pairs }
      } = await client?.frontierServiceListProjectUsers(projectId, {
        withRoles: true
      });
      setMembers(users);
      setMemberRoles(
        role_pairs.reduce((previous: any, mr: any) => {
          return { ...previous, [mr.user_id]: mr.roles };
        }, {})
      );
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message
      });
    } finally {
      setIsMembersLoading(false);
    }
  }, [client, isDeleteRoute, organization?.id, projectId]);

  const getProjectDetails = useCallback(async () => {
    if (!organization?.id || !projectId || isDeleteRoute) return;
    try {
      setIsProjectLoading(true);
      const {
        // @ts-ignore
        data: { project }
      } = await client?.frontierServiceGetProject(projectId);
      setProject(project);
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message
      });
    } finally {
      setIsProjectLoading(false);
    }
  }, [client, isDeleteRoute, organization?.id, projectId]);

  useEffect(() => {
    getProjectDetails();
    getProjectMembers();
    getProjectTeams();
  }, [getProjectDetails, getProjectMembers, getProjectTeams]);

  const isLoading = isProjectLoading || isTeamsLoading || isMembersLoading;

  const refetchTeamAndMembers = useCallback(() => {
    getProjectMembers();
    getProjectTeams();
  }, [getProjectMembers, getProjectTeams]);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Image
          alt="back-icon"
          style={{ cursor: 'pointer' }}
          // @ts-ignore
          src={backIcon}
          onClick={() => navigate({ to: '/projects' })}
        />
        <Text size={6}>Projects</Text>
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
          <General
            organization={organization}
            project={project}
            isLoading={isProjectLoading}
          />
        </Tabs.Content>
        <Tabs.Content value="members">
          <Members
            members={members}
            memberRoles={memberRoles}
            isLoading={isLoading}
            teams={teams}
            refetch={refetchTeamAndMembers}
          />
        </Tabs.Content>
      </Tabs>
      <Outlet />
    </Flex>
  );
};
