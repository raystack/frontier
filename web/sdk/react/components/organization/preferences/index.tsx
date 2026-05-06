'use client';

import { useRouteContext } from '@tanstack/react-router';
import { PreferencesPage } from '~/react/views/preferences';
import { RouterContext } from '../routes';

export default function UserPreferences() {
  const { theme, onThemeChange } = useRouteContext({ from: '__root__' }) as RouterContext;
  return <PreferencesPage theme={theme} onThemeChange={onThemeChange} />;
}
