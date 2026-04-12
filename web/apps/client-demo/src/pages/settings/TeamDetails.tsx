import { useParams, useNavigate } from 'react-router-dom';
import { TeamDetailsView } from '@raystack/frontier/react';

export default function TeamDetails() {
  const { orgId, teamId } = useParams<{ orgId: string; teamId: string }>();
  const navigate = useNavigate();

  if (!teamId) return null;

  return (
    <TeamDetailsView
      teamId={teamId}
      onNavigateToTeams={() => navigate(`/${orgId}/settings/teams`)}
      onDeleteSuccess={() => navigate(`/${orgId}/settings/teams`)}
    />
  );
}
