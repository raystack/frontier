import { useParams, useNavigate } from 'react-router-dom';
import { TeamsView } from '@raystack/frontier/react';

export default function Teams() {
  const { orgId } = useParams<{ orgId: string }>();
  const navigate = useNavigate();

  return (
    <TeamsView
      onTeamClick={(teamId) => navigate(`/${orgId}/settings/teams/${teamId}`)}
    />
  );
}
