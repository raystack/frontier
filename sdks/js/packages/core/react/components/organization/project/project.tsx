import { Tabs, Image, Text, toast, Flex } from '@raystack/apsara/v1';
import {
  Outlet,
  useLocation,
  useNavigate,
  useParams,
  useRouterState
} from '@tanstack/react-router';
import { useCallback, useEffect, useMemo, useState } from 'react';
import backIcon from '~/react/assets/chevron-left.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import type {
  V1Beta1Group,
  V1Beta1Project,
  V1Beta1Role,
  V1Beta1User
} from '~/src';
import type { Role } from '~/src/types';
import { PERMISSIONS } from '~/utils';
import { General } from './general';
import { Members } from './members';
import styles from './project.module.css';

export const ProjectPage = () => {
  let { projectId } = useParams({ from: '/projects/$projectId' });
  const [isProjectLoading, setIsProjectLoading] = useState(false);
  const [isProjectRoleLoading, setIsProjectRoleLoading] = useState(false);
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [project, setProject] = useState<V1Beta1Project>();
  const [members, setMembers] = useState<V1Beta1User[]>([]);
  const [memberRoles, setMemberRoles] = useState<Record<string, Role[]>>({});
  const [groupRoles, setGroupRoles] = useState<Record<string, Role[]>>({});
  const [isMembersLoading, setIsMembersLoading] = useState(false);

  const [teams, setTeams] = useState<V1Beta1Group[]>([]);
  const [isTeamsLoading, setIsTeamsLoading] = useState(false);

  const { client, activeOrganization: organization } = useFrontier();
  let navigate = useNavigate({ from: '/projects/$projectId' });
  const routeState = useRouterState();

  const location = useLocation();
  const refetch = location?.state?.refetch;

  const isDeleteRoute = useMemo(() => {
    return routeState.matches.some(
      route => route.routeId === '/projects/$projectId/delete'
    );
  }, [routeState.matches]);

  const getProjectTeams = useCallback(async () => {
    if (!organization?.id || !projectId || isDeleteRoute) return;
    try {
      setIsTeamsLoading(true);
      const resp = await client?.frontierServiceListProjectGroups(projectId, {
        with_roles: true
      });
      const groups = resp?.data?.groups || [];
      const role_pairs = resp?.data?.role_pairs || [];
      setTeams(groups);
      setGroupRoles(
        role_pairs.reduce((previous: any, gr: any) => {
          return { ...previous, [gr.group_id]: gr.roles };
        }, {})
      );
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
      const resp = await client?.frontierServiceListProjectUsers(projectId, {
        with_roles: true
      });
      const users = resp?.data?.users || [];
      const role_pairs = resp?.data?.role_pairs || [];
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
      const resp = await client?.frontierServiceGetProject(projectId);
      const project = resp?.data?.project;
      setProject(project);
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message
      });
    } finally {
      setIsProjectLoading(false);
    }
  }, [client, isDeleteRoute, organization?.id, projectId]);

  const getProjectRoles = useCallback(async () => {
    if (!organization?.id || !projectId || isDeleteRoute) return;
    try {
      setIsProjectRoleLoading(true);
      const resp = await client?.frontierServiceListRoles({
        state: 'enabled',
        scopes: [PERMISSIONS.ProjectNamespace]
      });
      const roles = resp?.data?.roles || [];
      setRoles(roles);
    } catch (error: any) {
      toast.error('Something went wrong', {
        description: error?.message
      });
    } finally {
      setIsProjectRoleLoading(false);
    }
  }, [client, isDeleteRoute, organization?.id, projectId]);

  useEffect(() => {
    getProjectDetails();
    getProjectMembers();
    getProjectTeams();
    getProjectRoles();
  }, [
    getProjectDetails,
    getProjectMembers,
    getProjectTeams,
    getProjectRoles,
    refetch
  ]);

  const isLoading =
    isProjectLoading ||
    isTeamsLoading ||
    isMembersLoading ||
    isProjectRoleLoading;

  const refetchTeamAndMembers = useCallback(() => {
    getProjectMembers();
    getProjectTeams();
  }, [getProjectMembers, getProjectTeams]);

  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex className={styles.header}>
        <Image
          alt="back-icon"
          style={{ cursor: 'pointer' }}
          src={backIcon as unknown as string}
          onClick={() => navigate({ to: '/projects' })}
          data-test-id="frontier-sdk-projects-page-back-link"
        />
        <Text size="large">Projects</Text>
      </Flex>
      <Tabs.Root defaultValue="general" className={styles.container}>
        <Tabs.List>
          <Tabs.Trigger value="general">General</Tabs.Trigger>
          <Tabs.Trigger value="members">Members</Tabs.Trigger>
        </Tabs.List>
        <Tabs.Content value="general">
          <General
            organization={organization}
            project={project}
            isLoading={isProjectLoading}
          />
        </Tabs.Content>
        <Tabs.Content value="members" className={styles.tabContent}>
          <Members
            members={members}
            memberRoles={memberRoles}
            groupRoles={groupRoles}
            isLoading={isLoading}
            teams={teams}
            roles={roles}
            refetch={refetchTeamAndMembers}
          />
        </Tabs.Content>
      </Tabs.Root>
      <Outlet />
    </Flex>
  );
};
