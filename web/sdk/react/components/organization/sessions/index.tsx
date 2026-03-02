'use client';

import { useRouteContext } from '@tanstack/react-router';
import { SessionsPage as SessionsPageView } from '~/react/views/sessions';

export const SessionsPage = () => {
  const { onLogout } = useRouteContext({ from: '__root__' }) as { onLogout?: () => void };
  return <SessionsPageView onLogout={onLogout} />;
};
