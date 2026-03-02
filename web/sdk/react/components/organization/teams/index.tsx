import { useNavigate } from '@tanstack/react-router';
import { TeamsListPage } from '~/react/views/teams';

export default function WorkspaceTeams() {
  const navigate = useNavigate({ from: '/teams' });

  return <TeamsListPage
    onTeamClick={(teamId: string) =>
      navigate({ to: '/teams/$teamId', params: { teamId } })
    }
  />;
}