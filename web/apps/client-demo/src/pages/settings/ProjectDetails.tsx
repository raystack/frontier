import { useParams, useNavigate } from 'react-router-dom';
import { ProjectDetailsView } from '@raystack/frontier/react';

export default function ProjectDetails() {
  const { orgId, projectId } = useParams<{ orgId: string; projectId: string }>();
  const navigate = useNavigate();

  if (!projectId) return null;

  return (
    <ProjectDetailsView
      projectId={projectId}
      onNavigateToProjects={() => navigate(`/${orgId}/settings/projects`)}
      onDeleteSuccess={() => navigate(`/${orgId}/settings/projects`)}
    />
  );
}
