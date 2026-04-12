'use client';

import { useMemo } from 'react';
import { Text } from '@raystack/apsara-v1';
import { useQuery } from '@connectrpc/connect-query';
import { create } from '@bufbuild/protobuf';
import {
  FrontierServiceQueries,
  ListServiceUserProjectsRequestSchema
} from '@raystack/proton/frontier';

interface ProjectsCellProps {
  serviceUserId: string;
  orgId: string;
}

export function ProjectsCell({ serviceUserId, orgId }: ProjectsCellProps) {
  const { data } = useQuery(
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

  return (
    <Text
      size="small"
      style={{
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap'
      }}
    >
      {projectNames || '-'}
    </Text>
  );
}
