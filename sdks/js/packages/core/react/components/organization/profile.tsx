import {
  RouterProvider,
  createMemoryHistory,
  createRouter
} from '@tanstack/react-router';
import { OrganizationProfileProps, routeTree } from './routes';

const router = createRouter({
  routeTree,
  context: {
    organizationId: '',
    showBilling: false,
    showTokens: false,
    showPreferences: false
  }
});

export const OrganizationProfile = ({
  organizationId,
  defaultRoute = '/',
  showBilling = false,
  showTokens = false,
  showPreferences = false,
  hideToast = false
}: OrganizationProfileProps) => {
  const memoryHistory = createMemoryHistory({
    initialEntries: [defaultRoute]
  });

  const memoryRouter = createRouter({
    routeTree,
    history: memoryHistory,
    context: {
      organizationId,
      showBilling,
      showTokens,
      hideToast,
      showPreferences
    }
  });
  return <RouterProvider router={memoryRouter} />;
};

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
