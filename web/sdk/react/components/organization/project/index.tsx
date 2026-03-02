'use client';

import { useNavigate } from '@tanstack/react-router';
import { ProjectsListPage } from '~/react/views/projects';

export default function WorkspaceProjects() {
  const navigate = useNavigate({ from: '/projects' });

  return (
    <ProjectsListPage
      onProjectClick={(projectId) =>
        navigate({ to: '/projects/$projectId', params: { projectId } })
      }
    />
  );
}
