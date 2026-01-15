import { useEffect } from 'react';
import type { AdminConfig } from '../types';

export function AdminPageTitle({
  title,
  appName,
  config
}: {
  title?: string;
  appName?: string;
  config?: AdminConfig;
}) {
  const titleAppName = appName || config?.title || 'Frontier Admin';
  const fullTitle = title ? `${title} | ${titleAppName}` : titleAppName;

  useEffect(() => {
    document.title = fullTitle;
    return () => {
      document.title = titleAppName;
    };
  }, [fullTitle, titleAppName]);

  return null;
}

