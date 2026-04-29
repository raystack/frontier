'use client';

import { useMemo } from 'react';
import { Skeleton, Text, Tooltip } from '@raystack/apsara-v1';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  ListServiceUserProjectsRequestSchema
} from '@raystack/proton/frontier';
import styles from './projects-cell.module.css';

interface ProjectsCellProps {
  serviceUserId: string;
  orgId: string;
}

export function ProjectsCell({ serviceUserId, orgId }: ProjectsCellProps) {
  const { data, isLoading } = useQuery(
    FrontierServiceQueries.listServiceUserProjects,
    create(ListServiceUserProjectsRequestSchema, {
      id: serviceUserId,
      orgId,
      withPermissions: []
    }),
    {
      enabled: Boolean(serviceUserId) && Boolean(orgId)
    }
  );

  const projectNames = useMemo(() => {
    const projects = data?.projects ?? [];
    return projects.map(p => p.title).join(', ');
  }, [data]);

  if (isLoading) {
    return <Skeleton height="16px" width="200px" />;
  }

  if (!projectNames) {
    return <Text size="small">-</Text>;
  }

  return (
    <Tooltip>
      <Tooltip.Trigger
        render={<Text size="small" className={styles.text} />}
      >
        {projectNames}
      </Tooltip.Trigger>
      <Tooltip.Content className={styles.tooltipContent}>
        {projectNames}
      </Tooltip.Content>
    </Tooltip>
  );
}
