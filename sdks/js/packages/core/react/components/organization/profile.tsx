import {
  RouterProvider,
  createMemoryHistory,
  createRouter
} from '@tanstack/react-router';
import { getCustomRoutes, OrganizationProfileProps, routeTree } from './routes';

const router = createRouter({
  routeTree,
  context: {
    organizationId: '',
    showBilling: false,
    showTokens: false,
    showPreferences: false,
    customRoutes: { Organization: [], User: [] }
  }
});

export const OrganizationProfile = ({
  organizationId,
  defaultRoute = '/',
  showBilling = false,
  showTokens = false,
  showPreferences = false,
  hideToast = false,
  customComponents
}: OrganizationProfileProps) => {
  const memoryHistory = createMemoryHistory({
    initialEntries: [defaultRoute]
  });

  const customRoutes = getCustomRoutes(customComponents);

  const memoryRouter = createRouter({
    routeTree,
    history: memoryHistory,
    context: {
      organizationId,
      showBilling,
      showTokens,
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
