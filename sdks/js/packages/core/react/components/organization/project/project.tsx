import { Flex, Image, Text } from '@raystack/apsara';

import { Tabs } from '@raystack/apsara';
import {
  Outlet,
  useNavigate,
  useParams,
  useRouterState
} from '@tanstack/react-router';
import { useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';
import backIcon from '~/react/assets/chevron-left.svg';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Project, V1Beta1User } from '~/src';
import { Role } from '~/src/types';
import { styles } from '../styles';
import { General } from './general';
import { Members } from './members';

export const ProjectPage = () => {
  let { projectId } = useParams({ from: '/projects/$projectId' });
  const [isLoading, setIsLoading] = useState(false);
  const [project, setProject] = useState<V1Beta1Project>();
  const [members, setMembers] = useState<V1Beta1User[]>([]);
  const [memberRoles, setMemberRoles] = useState<Record<string, Role[]>>({});
  const { client, activeOrganization: organization } = useFrontier();
  let navigate = useNavigate({ from: '/projects/$projectId' });
  const routeState = useRouterState();

  const isDeleteRoute = useMemo(() => {
    return routeState.matches.some(
      route => route.routeId === '/projects/$projectId/delete'
    );
  }, [routeState.matches]);

  useEffect(() => {
    async function getProjectDetails() {
      if (!organization?.id || !projectId || isDeleteRoute) return;

      try {
        setIsLoading(true);
        const {
          // @ts-ignore
          data: { project }
        } = await client?.frontierServiceGetProject(projectId);

        const {
          // @ts-ignore
          data: { users, role_pairs }
        } = await client?.frontierServiceListProjectUsers(projectId, {
          withRoles: true
        });
        setProject(project);
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
        setIsLoading(false);
      }
    }
    getProjectDetails();
  }, [
    client,
    organization?.id,
    projectId,
    routeState.location.key,
    isDeleteRoute
  ]);

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
            isLoading={isLoading}
          />
        </Tabs.Content>
        <Tabs.Content value="members">
          <Members
            members={members}
            memberRoles={memberRoles}
            isLoading={isLoading}
          />
        </Tabs.Content>
      </Tabs>
      <Outlet />
    </Flex>
  );
};
