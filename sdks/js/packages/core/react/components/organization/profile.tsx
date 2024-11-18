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
  customScreens = []
}: OrganizationProfileProps) => {
  const memoryHistory = createMemoryHistory({
    initialEntries: [defaultRoute]
  });

  const customRoutes = getCustomRoutes(customScreens);

  const routeTree = getRootTree({ customScreens });

  const memoryRouter = createRouter({
    routeTree,
    history: memoryHistory,
    context: {
      organizationId,
      showBilling,
      showTokens,
      showAPIKeys,
      hideToast,
      showPreferences,
      customRoutes
    }
  });
  return <RouterProvider router={memoryRouter} />;
};

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
