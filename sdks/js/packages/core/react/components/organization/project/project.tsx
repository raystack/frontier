import { Flex, Text, Image } from '@raystack/apsara';

import { Tabs } from '@raystack/apsara';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { V1Beta1Project, V1Beta1User } from '~/src';
import { styles } from '../styles';
import { General } from './general';
import { Members } from './members';
import backIcon from '~/react/assets/chevron-left.svg';

export const ProjectPage = () => {
  let { projectId } = useParams({ from: '/projects/$projectId' });
  const [project, setProject] = useState<V1Beta1Project>();
  const [members, setMembers] = useState<V1Beta1User[]>([]);
  const { client, activeOrganization: organization } = useFrontier();
  let navigate = useNavigate({ from: '/projects/$projectId' });

  useEffect(() => {
    async function getProjectDetails() {
      if (!organization?.id || !projectId) return;

      try {
        const {
          // @ts-ignore
          data: { project }
        } = await client?.frontierServiceGetProject(projectId);

        const {
          // @ts-ignore
          data: { users }
        } = await client?.frontierServiceListProjectUsers(projectId, {
          withRoles: true
        });
        setProject(project);
        setMembers(users);
      } catch ({ error }: any) {
        toast.error('Something went wrong', {
          description: error.message
        });
      }
    }
    getProjectDetails();
  }, [client, organization?.id, projectId]);

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
          <General organization={organization} project={project} />
        </Tabs.Content>
        <Tabs.Content value="members">
          <Members members={members} />
        </Tabs.Content>
      </Tabs>
    </Flex>
  );
};
