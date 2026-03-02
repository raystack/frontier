'use client';

import { useNavigate, useParams } from '@tanstack/react-router';
import { TeamDetailPage } from '~/react/views/teams';

export const TeamPage = () => {
  const { teamId } = useParams({ from: '/teams/$teamId' });
  const navigate = useNavigate({ from: '/teams/$teamId' });

  return (
    <TeamDetailPage
      teamId={teamId}
      onBack={() => navigate({ to: '/teams' })}
    />
  );
};
