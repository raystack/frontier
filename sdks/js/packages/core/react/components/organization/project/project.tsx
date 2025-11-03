import { Tabs, Image, Text, toast, Flex } from '@raystack/apsara';
import {
  Outlet,
  useNavigate,
  useParams,
  useRouterState
} from '@tanstack/react-router';
import { useCallback, useEffect, useMemo } from 'react';
import backIcon from '~/react/assets/chevron-left.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries,
  ListProjectGroupsRequestSchema,
  ListProjectUsersRequestSchema,
  GetProjectRequestSchema,
  ListRolesRequestSchema,
  type Role as ProtoRole,
  type Organization
} from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { PERMISSIONS } from '~/utils';
import { General } from './general';
import { Members } from './members';
import styles from './project.module.css';

interface ProjectGroupRolePair {
  groupId?: string;
  group_id?: string;
  roles: ProtoRole[];
}

interface ProjectUserRolePair {
  userId?: string;
  user_id?: string;
  roles: ProtoRole[];
}

export const ProjectPage = () => {
  let { projectId } = useParams({ from: '/projects/$projectId' });

  const { activeOrganization: organization } = useFrontier();
  let navigate = useNavigate({ from: '/projects/$projectId' });
  const routeState = useRouterState();

  const isDeleteRoute = useMemo(() => {
    return routeState.matches.some(
      route => route.routeId === '/projects/$projectId/delete'
    );
  }, [routeState.matches]);

  const {
    data: projectGroups = { groups: [], groupRoles: {} },
    isLoading: isTeamsLoading,
    error: projectGroupsError,
    refetch: refetchProjectGroups
  } = useQuery(
    FrontierServiceQueries.listProjectGroups,
    create(ListProjectGroupsRequestSchema, {
      id: projectId || '',
      withRoles: true
    }),
    {
      enabled: !!organization?.id && !!projectId && !isDeleteRoute,
      select: d => ({
        groups: d?.groups ?? [],
        groupRoles: (d?.rolePairs ?? []).reduce((acc: Record<string, ProtoRole[]>, gr: ProjectGroupRolePair) => {
          const key = gr.groupId ?? gr.group_id;
          if (key) acc[key] = gr.roles;
          return acc;
        }, {})
      })
    }
  );

  useEffect(() => {
    if (projectGroupsError) {
      toast.error('Something went wrong', {
        description: projectGroupsError.message
      });
    }
  }, [projectGroupsError]);

  const {
    data: projectUsers = { users: [], memberRoles: {} },
    isLoading: isMembersLoadingQuery,
    refetch: refetchProjectUsers
  } = useQuery(
    FrontierServiceQueries.listProjectUsers,
    create(ListProjectUsersRequestSchema, {
      id: projectId || '',
      withRoles: true
    }),
    {
      enabled: !!organization?.id && !!projectId && !isDeleteRoute,
      select: d => ({
        users: d?.users ?? [],
        memberRoles: (d?.rolePairs ?? []).reduce((acc: Record<string, ProtoRole[]>, mr: ProjectUserRolePair) => {
          const key = mr.userId ?? mr.user_id;
          if (key) acc[key] = mr.roles;
          return acc;
        }, {})
      })
    }
  );

  const {
    data: project,
    isLoading: isProjectLoadingQuery,
    error: projectError
  } = useQuery(
    FrontierServiceQueries.getProject,
    create(GetProjectRequestSchema, { id: projectId || '' }),
    {
      enabled: !!organization?.id && !!projectId && !isDeleteRoute,
      select: (d) => d?.project
    }
  );

  useEffect(() => {
    if (projectError) {
      toast.error('Something went wrong', { description: projectError.message });
    }
  }, [projectError]);

  const {
    data: roles = [],
    isLoading: isProjectRoleLoadingQuery,
    error: rolesError
  } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, {
      state: 'enabled',
      scopes: [PERMISSIONS.ProjectNamespace]
    }),
    {
      enabled: !!organization?.id && !!projectId && !isDeleteRoute,
      select: d => (d?.roles ?? [])
    }
  );

  useEffect(() => {
    if (rolesError) {
      toast.error('Something went wrong', { description: rolesError.message });
    }
  }, [rolesError]);

  const isLoading =
    isProjectLoadingQuery ||
    isTeamsLoading ||
    isMembersLoadingQuery ||
    isProjectRoleLoadingQuery;

  const refetchTeamAndMembers = useCallback(() => {
    refetchProjectUsers();
    refetchProjectGroups();
  }, [refetchProjectUsers, refetchProjectGroups]);

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
      <Tabs defaultValue="general" className={styles.container}>
        <Tabs.List>
          <Tabs.Trigger value="general">General</Tabs.Trigger>
          <Tabs.Trigger value="members">Members</Tabs.Trigger>
        </Tabs.List>
        <Tabs.Content value="general">
          <General
            organization={organization as Organization}
            project={project}
            isLoading={isProjectLoadingQuery}
          />
        </Tabs.Content>
        <Tabs.Content value="members" className={styles.tabContent}>
          <Members
            members={projectUsers.users}
            memberRoles={projectUsers.memberRoles}
            groupRoles={projectGroups.groupRoles}
            isLoading={isLoading}
            teams={projectGroups.groups}
            roles={roles}
            refetch={refetchTeamAndMembers}
          />
        </Tabs.Content>
      </Tabs>
      <Outlet />
    </Flex>
  );
};
