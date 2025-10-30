import { Tabs, Image, Text, toast, Flex } from '@raystack/apsara';
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
import { useQuery } from '@connectrpc/connect-query';
import { FrontierServiceQueries, ListProjectGroupsRequestSchema, ListProjectUsersRequestSchema, GetProjectRequestSchema, ListRolesRequestSchema, type Role as ProtoRole, type Group, type User } from '@raystack/proton/frontier';
import { create } from '@bufbuild/protobuf';
import { PERMISSIONS } from '~/utils';
import { General } from './general';
import { Members } from './members';
import styles from './project.module.css';

export const ProjectPage = () => {
  let { projectId } = useParams({ from: '/projects/$projectId' });
  const [isProjectLoading, setIsProjectLoading] = useState(false);
  const [isProjectRoleLoading, setIsProjectRoleLoading] = useState(false);
  const [roles, setRoles] = useState<ProtoRole[]>([]);
  const [project, setProject] = useState<any>();
  const [members, setMembers] = useState<User[]>([]);
  const [memberRoles, setMemberRoles] = useState<Record<string, ProtoRole[]>>({});
  const [groupRoles, setGroupRoles] = useState<Record<string, ProtoRole[]>>({});
  const [isMembersLoading, setIsMembersLoading] = useState(false);

  const [teams, setTeams] = useState<Group[]>([]);

  const { activeOrganization: organization } = useFrontier();
  let navigate = useNavigate({ from: '/projects/$projectId' });
  const routeState = useRouterState();

  const location = useLocation();
  const refetch = location?.state?.refetch;

  const isDeleteRoute = useMemo(() => {
    return routeState.matches.some(
      route => route.routeId === '/projects/$projectId/delete'
    );
  }, [routeState.matches]);

  const {
    data: projectGroupsData,
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
        groupRoles: (d?.rolePairs ?? []).reduce((acc: Record<string, ProtoRole[]>, gr: any) => {
          const key = gr.groupId ?? gr.group_id;
          acc[key] = gr.roles as ProtoRole[];
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

  useEffect(() => {
    if (projectGroupsData) {
      setTeams(projectGroupsData.groups as Group[]);
      setGroupRoles(projectGroupsData.groupRoles as Record<string, ProtoRole[]>);
    }
  }, [projectGroupsData]);

  const {
    data: projectUsersData,
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
        memberRoles: (d?.rolePairs ?? []).reduce((acc: Record<string, ProtoRole[]>, mr: any) => {
          const key = mr.userId ?? mr.user_id;
          acc[key] = mr.roles as ProtoRole[];
          return acc;
        }, {})
      })
    }
  );

  useEffect(() => {
    if (projectUsersData) {
      setMembers(projectUsersData.users as User[]);
      setMemberRoles(projectUsersData.memberRoles as Record<string, ProtoRole[]>);
    }
  }, [projectUsersData]);

  useEffect(() => {
    setIsMembersLoading(isMembersLoadingQuery);
  }, [isMembersLoadingQuery]);

  const { data: projectData, isLoading: isProjectLoadingQuery, error: projectError } = useQuery(
    FrontierServiceQueries.getProject,
    create(GetProjectRequestSchema, { id: projectId || '' }),
    {
      enabled: !!organization?.id && !!projectId && !isDeleteRoute
    }
  );

  useEffect(() => {
    if (projectData?.project) {
      setProject(projectData.project as unknown);
    }
  }, [projectData]);

  useEffect(() => {
    if (projectError) {
      toast.error('Something went wrong', { description: projectError.message });
    }
  }, [projectError]);

  useEffect(() => {
    setIsProjectLoading(isProjectLoadingQuery);
  }, [isProjectLoadingQuery]);

  const { data: rolesData, isLoading: isProjectRoleLoadingQuery, error: rolesError } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, {
      state: 'enabled',
      scopes: [PERMISSIONS.ProjectNamespace]
    }),
    { enabled: !!organization?.id && !!projectId && !isDeleteRoute }
  );

  useEffect(() => {
    if (rolesError) {
      toast.error('Something went wrong', { description: rolesError.message });
    }
  }, [rolesError]);

  useEffect(() => {
    setRoles((rolesData?.roles || []) as ProtoRole[]);
  }, [rolesData]);

  useEffect(() => {
    setIsProjectRoleLoading(isProjectRoleLoadingQuery);
  }, [isProjectRoleLoadingQuery]);

  useEffect(() => {
    // roles fetched via query above
  }, [refetch]);

  useEffect(() => {
    if (refetch) {
      refetchProjectUsers();
      refetchProjectGroups();
    }
  }, [refetch, refetchProjectUsers, refetchProjectGroups]);

  const isLoading =
    isProjectLoading ||
    isTeamsLoading ||
    isMembersLoading ||
    isProjectRoleLoading;

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
            roles={roles as ProtoRole[]}
            refetch={refetchTeamAndMembers}
          />
        </Tabs.Content>
      </Tabs>
      <Outlet />
    </Flex>
  );
};
