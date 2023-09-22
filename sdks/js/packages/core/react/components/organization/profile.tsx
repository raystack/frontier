import { Flex, ThemeProvider } from '@raystack/apsara';
import { useCallback, useEffect, useState } from 'react';
import {
  Outlet,
  RouterProvider,
  Router,
  Route,
  RootRoute,
  createMemoryHistory
} from '@tanstack/react-router';
import { Toaster } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import Domain from './domain';
import { AddDomain } from './domain/add-domain';
import { VerifyDomain } from './domain/verify-domain';
import GeneralSetting from './general';
import { DeleteOrganization } from './general/delete';
import WorkspaceMembers from './members';
import { InviteMember } from './members/invite';
import UserPreferences from './preferences';

import { default as WorkspaceProjects } from './project';
import { AddProject } from './project/add';
import { DeleteProject } from './project/delete';
import { ProjectPage } from './project/project';
import { InviteProjectTeam } from './project/members/invite';

import WorkspaceSecurity from './security';
import { Sidebar } from './sidebar';
import WorkspaceTeams from './teams';
import { AddTeam } from './teams/add';
import { DeleteTeam } from './teams/delete';
import { TeamPage } from './teams/team';
import { UserSetting } from './user';
import { SkeletonTheme } from 'react-loading-skeleton';

interface OrganizationProfileProps {
  organizationId: string;
  defaultRoute?: string;
}

const rootRoute = new RootRoute({
  component: () => {
    return (
      <ThemeProvider>
        <SkeletonTheme
          highlightColor="var(--background-base)"
          baseColor="var(--background-base-hover)"
        >
          <Toaster richColors />
          <Flex style={{ width: '100%', height: '100%' }}>
            <Sidebar />
            <Outlet />
          </Flex>
        </SkeletonTheme>
      </ThemeProvider>
    );
  }
});

const indexRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/',
  component: GeneralSetting
});

const deleteOrgRoute = new Route({
  getParentRoute: () => indexRoute,
  path: '/delete',
  component: DeleteOrganization
});

const securityRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/security',
  component: WorkspaceSecurity
});

const membersRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/members',
  component: WorkspaceMembers
});

const inviteMemberRoute = new Route({
  getParentRoute: () => membersRoute,
  path: '/modal',
  component: InviteMember
});

const teamsRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/teams',
  component: WorkspaceTeams
});

const addTeamRoute = new Route({
  getParentRoute: () => teamsRoute,
  path: '/modal',
  component: AddTeam
});

const domainsRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/domains',
  component: Domain
});

const verifyDomainRoute = new Route({
  getParentRoute: () => domainsRoute,
  path: '/$domainId/verify',
  component: VerifyDomain
});

const addDomainRoute = new Route({
  getParentRoute: () => domainsRoute,
  path: '/modal',
  component: AddDomain
});

const teamRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/teams/$teamId',
  component: TeamPage
});

const deleteTeamRoute = new Route({
  getParentRoute: () => teamRoute,
  path: '/delete',
  component: DeleteTeam
});

const projectsRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/projects',
  component: WorkspaceProjects
});

const addProjectRoute = new Route({
  getParentRoute: () => projectsRoute,
  path: '/modal',
  component: AddProject
});

const projectPageRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/projects/$projectId',
  component: ProjectPage
});

const deleteProjectRoute = new Route({
  getParentRoute: () => projectPageRoute,
  path: '/delete',
  component: DeleteProject
});

const projectTeamInviteRoute = new Route({
  getParentRoute: () => projectPageRoute,
  path: '/invite',
  component: InviteProjectTeam
});

const profileRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/profile',
  component: UserSetting
});

const preferencesRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/preferences',
  component: UserPreferences
});

const routeTree = rootRoute.addChildren([
  indexRoute.addChildren([deleteOrgRoute]),
  securityRoute,
  membersRoute.addChildren([inviteMemberRoute]),
  teamsRoute.addChildren([addTeamRoute]),
  domainsRoute.addChildren([addDomainRoute, verifyDomainRoute]),
  teamRoute.addChildren([deleteTeamRoute]),
  projectsRoute.addChildren([addProjectRoute]),
  projectPageRoute.addChildren([deleteProjectRoute, projectTeamInviteRoute]),
  profileRoute,
  preferencesRoute
]);

const router = new Router({ routeTree });

export const OrganizationProfile = ({
  organizationId,
  defaultRoute = '/'
}: OrganizationProfileProps) => {
  const { client, setActiveOrganization, setIsActiveOrganizationLoading } =
    useFrontier();

  const fetchOrganization = useCallback(async () => {
    try {
      setIsActiveOrganizationLoading(true);
      const {
        // @ts-ignore
        data: { organization }
      } = await client?.frontierServiceGetOrganization(organizationId);
      setActiveOrganization(organization);
    } catch (err) {
      console.error(err);
    } finally {
      setIsActiveOrganizationLoading(false);
    }
  }, [
    client,
    organizationId,
    setActiveOrganization,
    setIsActiveOrganizationLoading
  ]);

  useEffect(() => {
    if (organizationId) {
      fetchOrganization();
    } else {
      setActiveOrganization(undefined);
    }
  }, [organizationId, fetchOrganization, setActiveOrganization]);

  const memoryHistory = createMemoryHistory({
    initialEntries: [defaultRoute]
  });

  const memoryRouter = new Router({ routeTree, history: memoryHistory });

  return <RouterProvider router={memoryRouter} />;
};

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
