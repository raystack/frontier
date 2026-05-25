'use client';

import { useNavigate, useParams } from '@tanstack/react-router';
import { PATDetailsView } from '~/react/views-new/pat';

export function PatPage() {
  const { patId } = useParams({ from: '/pats/$patId' });
  const navigate = useNavigate({ from: '/pats/$patId' });

  return (
    <PATDetailsView
      patId={patId}
      onNavigateToPats={() => navigate({ to: '/pats' })}
      onDeleteSuccess={() => navigate({ to: '/pats' })}
    />
  );
}
