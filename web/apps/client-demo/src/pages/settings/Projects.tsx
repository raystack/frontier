import { useParams, useNavigate } from 'react-router-dom';
import { ProjectsView } from '@raystack/frontier/react';

export default function Projects() {
  const { orgId } = useParams<{ orgId: string }>();
  const navigate = useNavigate();

  return (
    <ProjectsView
      onProjectClick={(projectId) => navigate(`/${orgId}/settings/projects/${projectId}`)}
    />
  );
}
