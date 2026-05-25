'use client';

import { useNavigate } from '@tanstack/react-router';
import { PatsView } from '~/react/views-new/pat';

export default function WorkspacePats() {
  const navigate = useNavigate({ from: '/pats' });

  return (
    <PatsView
      onPATClick={(patId: string) =>
        navigate({ to: '/pats/$patId', params: { patId } })
      }
    />
  );
}
