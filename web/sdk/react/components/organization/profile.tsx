import { useMemo } from 'react';
import {
  RouterProvider,
  createMemoryHistory,
  createRouter
} from '@tanstack/react-router';
import {
  getCustomRoutes,
  getRootTree,
  OrganizationProfileProps
} from './routes';

const router = createRouter({
  routeTree: getRootTree({}),
  context: {
    organizationId: '',
    showBilling: false,
    showTokens: false,
    showAPIKeys: false,
    showPreferences: false,
    customRoutes: { Organization: [], User: [] }
  }
});

export const OrganizationProfile = ({
  organizationId,
  defaultRoute = '/',
  showBilling = false,
  showTokens = false,
  showAPIKeys = false,
  showPreferences = false,
  hideToast = false,
  customScreens = [],
  onLogout = () => { },
  theme,
  onThemeChange,
}: OrganizationProfileProps) => {
  const memoryHistory = createMemoryHistory({
    initialEntries: [defaultRoute]
  });

  const customRoutes = getCustomRoutes(customScreens);

  const routeTree = getRootTree({ customScreens });

  const memoryRouter = useMemo(
    () =>
      createRouter({
        routeTree,
        history: memoryHistory,
        context: {
          organizationId,
          showBilling,
          showTokens,
          showAPIKeys,
          hideToast,
          showPreferences,
          customRoutes,
          onLogout,
          theme,
          onThemeChange
        }
      }),
    // Router is created once; dynamic context flows through RouterProvider's
    // context prop so updates don't recreate the router and reset navigation.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  );

  return (
    <RouterProvider
      router={memoryRouter}
      context={{
        organizationId,
        showBilling,
        showTokens,
        showAPIKeys,
        hideToast,
        showPreferences,
        customRoutes,
        onLogout,
        theme,
        onThemeChange
      }}
    />
  );
};

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }

  interface HistoryState {
    refetch?: boolean;
    enableServiceUserTokensListFetch?: boolean;
  }
}
