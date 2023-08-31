import { Flex, ThemeProvider } from '@raystack/apsara';
import { useCallback, useEffect, useState } from 'react';
import {
  Outlet,
  RouterProvider,
  Router,
  Route,
  RootRoute
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
import WorkspaceSecurity from './security';
import { Sidebar } from './sidebar';
import WorkspaceTeams from './teams';
import { AddTeam } from './teams/add';
import { DeleteTeam } from './teams/delete';
import { TeamPage } from './teams/team';
import { UserSetting } from './user';
interface OrganizationProfileProps {
  organizationId: string;
  defaultRoute?: string;
}

const rootRoute = new RootRoute({
  component: () => {
    return (
      <ThemeProvider>
        <Toaster richColors />
        <Flex style={{ width: '100%', height: '100%' }}>
          <Sidebar />
          <Outlet />
        </Flex>
      </ThemeProvider>
    );
  }
});

export const OrganizationProfile = ({
  organizationId,
  defaultRoute = '/'
}: OrganizationProfileProps) => {
  const [organization, setOrganization] = useState();
  const { client, setActiveOrganization } = useFrontier();

  const fetchOrganization = useCallback(async () => {
    const {
      // @ts-ignore
      data: { organization }
    } = await client?.frontierServiceGetOrganization(organizationId);
    setOrganization(organization);
    setActiveOrganization(organization);
  }, [client, organizationId, setActiveOrganization]);

  useEffect(() => {
    if (organizationId) fetchOrganization();
  }, [organizationId, client, fetchOrganization]);

  const indexRoute = new Route({
    getParentRoute: () => rootRoute,
    path: '/',
    component: () => <GeneralSetting organization={organization} />
  });

  const deleteOrgRoute = new Route({
    getParentRoute: () => indexRoute,
    path: '/delete',
    component: () => <DeleteOrganization organization={organization} />
  });

  const securityRoute = new Route({
    getParentRoute: () => rootRoute,
    path: '/security',
    component: () => <WorkspaceSecurity organization={organization} />
  });

  const membersRoute = new Route({
    getParentRoute: () => rootRoute,
    path: '/members',
    component: () => <WorkspaceMembers organization={organization} />
  });

  const inviteMemberRoute = new Route({
    getParentRoute: () => membersRoute,
    path: '/modal',
    component: () => <InviteMember organization={organization} />
  });

  const teamsRoute = new Route({
    getParentRoute: () => rootRoute,
    path: '/members',
    component: () => <WorkspaceTeams organization={organization} />
  });

  const addTeamRoute = new Route({
    getParentRoute: () => teamsRoute,
    path: '/modal',
    component: () => <AddTeam organization={organization} />
  });

  const domainsRoute = new Route({
    getParentRoute: () => rootRoute,
    path: '/domains',
    component: () => <Domain organization={organization} />
  });

  const verifyDomainRoute = new Route({
    getParentRoute: () => domainsRoute,
    path: '/$domainId/verify',
    component: () => <VerifyDomain organization={organization} />
  });

  const addDomainRoute = new Route({
    getParentRoute: () => domainsRoute,
    path: '/modal',
    component: () => <AddDomain organization={organization} />
  });

  const teamRoute = new Route({
    getParentRoute: () => rootRoute,
    path: '/teams/$teamId',
    component: () => <TeamPage organization={organization} />
  });

  const deleteTeamRoute = new Route({
    getParentRoute: () => teamRoute,
    path: '/delete',
    component: () => <DeleteTeam organization={organization} />
  });

  const projectsRoute = new Route({
    getParentRoute: () => rootRoute,
    path: '/projects',
    component: () => <WorkspaceProjects organization={organization} />
  });

  const addProjectRoute = new Route({
    getParentRoute: () => projectsRoute,
    path: '/modal',
    component: () => <AddProject organization={organization} />
  });

  const projectPageRoute = new Route({
    getParentRoute: () => rootRoute,
    path: '/projects/$projectId',
    component: () => <ProjectPage organization={organization} />
  });

  const deleteProjectRoute = new Route({
    getParentRoute: () => projectPageRoute,
    path: '/delete',
    component: () => <DeleteProject organization={organization} />
  });

  const profileRoute = new Route({
    getParentRoute: () => rootRoute,
    path: '/profile',
    component: () => UserSetting
  });

  const perferencesRoute = new Route({
    getParentRoute: () => rootRoute,
    path: '/perferences',
    component: () => UserPreferences
  });

  const routeTree = rootRoute.addChildren([
    indexRoute.addChildren([deleteOrgRoute]),
    securityRoute,
    membersRoute.addChildren([inviteMemberRoute]),
    teamsRoute.addChildren([addTeamRoute]),
    domainsRoute.addChildren([addDomainRoute, verifyDomainRoute]),
    teamRoute.addChildren([deleteTeamRoute]),
    projectsRoute.addChildren([addProjectRoute]),
    projectPageRoute.addChildren([deleteProjectRoute]),
    profileRoute,
    perferencesRoute
  ]);
  const router = new Router({ routeTree });

  return <RouterProvider router={router} />;
};
