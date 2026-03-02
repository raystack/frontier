import { useEffect } from 'react';
import {
  Outlet,
  createRoute,
  createRootRouteWithContext,
  useRouteContext,
  RouteComponent
} from '@tanstack/react-router';
import { Flex, ToastContainer } from '@raystack/apsara';
import { useFrontier } from '~/react/contexts/FrontierContext';
import { useQuery, FrontierServiceQueries } from '~hooks';
import Domain from './domain';
import { AddDomain } from './domain/add-domain';
import { VerifyDomain } from './domain/verify-domain';
import GeneralSetting from './general';
import WorkspaceMembers from './members';
import UserPreferences from './preferences';

import { default as WorkspaceProjects } from './project';
import { ProjectPage } from './project/project';

import WorkspaceSecurity from './security';
import { Sidebar } from './sidebar';
import WorkspaceTeams from './teams';
import { TeamPage } from './teams/team';
import { UserSetting } from './user';
import { DeleteDomain } from './domain/delete';
import Billing from './billing';
import Tokens from './tokens';
import { AddTokens } from './tokens/add-tokens';
import { ConfirmCycleSwitch } from './billing/cycle-switch';
import Plans from './plans';
import ConfirmPlanChange from './plans/confirm-change';
import APIKeys from './api-keys';
import { AddServiceAccount } from './api-keys/add';
import ServiceUserPage from './api-keys/service-user';
import { DeleteServiceAccount } from './api-keys/delete';
import { DeleteServiceAccountKey } from './api-keys/service-user/delete';
import ManageServiceUserProjects from './api-keys/service-user/projects';
import { SessionsPage, RevokeSessionConfirm } from './sessions';
export interface CustomScreen {
  name: string;
  path: string;
  category: 'Organization' | 'User';
  component: RouteComponent;
}

export interface OrganizationProfileProps {
  organizationId: string;
  defaultRoute?: string;
  showBilling?: boolean;
  showTokens?: boolean;
  showAPIKeys?: boolean;
  showPreferences?: boolean;
  hideToast?: boolean;
  customScreens?: CustomScreen[];
  onLogout?: () => void;
}

export interface CustomRoutes {
  Organization: Pick<CustomScreen, 'name' | 'path'>[];
  User: Pick<CustomScreen, 'name' | 'path'>[];
}

type RouterContext = Pick<
  OrganizationProfileProps,
  | 'organizationId'
  | 'showBilling'
  | 'showTokens'
  | 'showAPIKeys'
  | 'hideToast'
  | 'showPreferences'
> & { customRoutes: CustomRoutes; onLogout?: () => void };

export function getCustomRoutes(customScreens: CustomScreen[] = []) {
  return (
    customScreens?.reduce(
      (acc: CustomRoutes, { name, category, path }) => {
        acc[category].push({ name, path });
        return acc;
      },
      { Organization: [], User: [] }
    ) || { Organization: [], User: [] }
  );
}

const RootRouter = () => {
  const { organizationId, hideToast } = useRouteContext({ from: '__root__' });
  const { setActiveOrganization, setIsActiveOrganizationLoading } =
    useFrontier();

  const {
    data: organizationData,
    isLoading,
    error
  } = useQuery(
    FrontierServiceQueries.getOrganization,
    { id: organizationId },
    { enabled: !!organizationId }
  );

  useEffect(() => {
    setIsActiveOrganizationLoading(isLoading);
  }, [isLoading, setIsActiveOrganizationLoading]);

  useEffect(() => {
    if (organizationId) {
      setActiveOrganization(organizationData?.organization);
    } else {
      setActiveOrganization(undefined);
    }
  }, [organizationId, organizationData?.organization, setActiveOrganization]);

  useEffect(() => {
    if (error) {
      console.error('Failed to fetch organization:', error);
    }
  }, [error]);

  const visibleToasts = hideToast ? 0 : 1;

  return (
    <>
      <ToastContainer richColors visibleToasts={visibleToasts} />
      <Flex style={{ width: '100%', height: '100%' }}>
        <Sidebar />
        <Outlet />
      </Flex>
    </>
  );
};

const rootRoute = createRootRouteWithContext<RouterContext>()({
  component: RootRouter
});
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: GeneralSetting
});

const securityRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/security',
  component: WorkspaceSecurity
});

const membersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/members',
  component: WorkspaceMembers
});

const teamsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/teams',
  component: WorkspaceTeams
});

const domainsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/domains',
  component: Domain
});

const verifyDomainRoute = createRoute({
  getParentRoute: () => domainsRoute,
  path: '/$domainId/verify',
  component: VerifyDomain
});

const deleteDomainRoute = createRoute({
  getParentRoute: () => domainsRoute,
  path: '/$domainId/delete',
  component: DeleteDomain
});

const addDomainRoute = createRoute({
  getParentRoute: () => domainsRoute,
  path: '/modal',
  component: AddDomain
});

const teamRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/teams/$teamId',
  component: TeamPage
});

const projectsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/projects',
  component: WorkspaceProjects
});

const projectPageRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/projects/$projectId',
  component: ProjectPage
});

const profileRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/profile',
  component: UserSetting
});

const preferencesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/preferences',
  component: UserPreferences
});

const billingRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/billing',
  component: Billing
});

const switchBillingCycleModalRoute = createRoute({
  getParentRoute: () => billingRoute,
  path: '/cycle-switch/$planId',
  component: ConfirmCycleSwitch
});

const plansRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/plans',
  component: Plans
});

const planDowngradeRoute = createRoute({
  getParentRoute: () => plansRoute,
  path: '/confirm-change/$planId',
  component: ConfirmPlanChange
});

const tokensRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/tokens',
  component: Tokens
});

const addTokensRoute = createRoute({
  getParentRoute: () => tokensRoute,
  path: '/modal',
  component: AddTokens
});

const apiKeysRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/api-keys',
  component: APIKeys
});

const addServiceAccountRoute = createRoute({
  getParentRoute: () => apiKeysRoute,
  path: '/add',
  component: AddServiceAccount
});

const deleteServiceAccountRoute = createRoute({
  getParentRoute: () => apiKeysRoute,
  path: '/$id/delete',
  component: DeleteServiceAccount
});

const serviceAccountRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/api-keys/$id',
  component: ServiceUserPage
});

const serviceAccountProjectRoute = createRoute({
  getParentRoute: () => serviceAccountRoute,
  path: '/projects',
  component: ManageServiceUserProjects
});

const deleteServiceAccountKeyRoute = createRoute({
  getParentRoute: () => serviceAccountRoute,
  path: '/key/$tokenId/delete',
  component: DeleteServiceAccountKey
});

const sessionsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/sessions',
  component: SessionsPage
});

const revokeSessionRoute = createRoute({
  getParentRoute: () => sessionsRoute,
  path: '/revoke',
  component: RevokeSessionConfirm,
  validateSearch: (search: Record<string, unknown>) => ({
    sessionId: search.sessionId as string | undefined,
  }),
});

interface getRootTreeOptions {
  customScreens?: CustomScreen[];
}

export function getRootTree({ customScreens = [] }: getRootTreeOptions) {
  return rootRoute.addChildren([
    indexRoute,
    securityRoute,
    sessionsRoute.addChildren([revokeSessionRoute]),
    membersRoute,
    teamsRoute,
    domainsRoute.addChildren([
      addDomainRoute,
      verifyDomainRoute,
      deleteDomainRoute
    ]),
    teamRoute,
    projectsRoute,
    projectPageRoute,
    profileRoute,
    preferencesRoute,
    billingRoute.addChildren([switchBillingCycleModalRoute]),
    plansRoute.addChildren([planDowngradeRoute]),
    tokensRoute.addChildren([addTokensRoute]),
    apiKeysRoute.addChildren([
      addServiceAccountRoute,
      deleteServiceAccountRoute
    ]),
    serviceAccountRoute.addChildren([
      deleteServiceAccountKeyRoute,
      serviceAccountProjectRoute
    ]),
    ...customScreens.map(cc =>
      createRoute({
        path: cc.path,
        component: cc.component,
        getParentRoute: () => rootRoute
      })
    )
  ]);
}
