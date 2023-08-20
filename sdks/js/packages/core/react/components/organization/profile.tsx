import { Flex, ThemeProvider } from '@raystack/apsara';
import { useCallback, useEffect, useState } from 'react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
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

export const OrganizationProfile = ({
  organizationId,
  defaultRoute = '/'
}: OrganizationProfileProps) => {
  const [organization, setOrganization] = useState();
  const { client } = useFrontier();

  const fetchOrganization = useCallback(async () => {
    const {
      // @ts-ignore
      data: { organization }
    } = await client?.frontierServiceGetOrganization(organizationId);
    setOrganization(organization);
  }, [client, organizationId]);

  useEffect(() => {
    if (organizationId) fetchOrganization();
  }, [organizationId, client, fetchOrganization]);

  return (
    // @ts-ignore
    <MemoryRouter initialEntries={[defaultRoute]}>
      <ThemeProvider>
        <Toaster richColors />
        <Flex style={{ width: '100%', height: '100%' }}>
          <Sidebar />
          <Flex style={{ flexGrow: '1', overflowY: 'auto' }}>
            <Routes>
              <Route
                path="/"
                element={<GeneralSetting organization={organization} />}
              >
                <Route
                  path="delete"
                  element={<DeleteOrganization organization={organization} />}
                />
              </Route>
              <Route path="security" element={<WorkspaceSecurity />} />
              <Route
                path="members"
                element={
                  <WorkspaceMembers
                    organization={organization}
                  ></WorkspaceMembers>
                }
              >
                <Route
                  path="modal"
                  element={<InviteMember organization={organization} />}
                />
              </Route>

              <Route
                path="teams"
                element={<WorkspaceTeams organization={organization} />}
              >
                <Route
                  path="modal"
                  element={<AddTeam organization={organization} />}
                />
              </Route>

              <Route
                path="domains"
                element={<Domain organization={organization} />}
              >
                <Route
                  path=":domainId/verify"
                  element={<VerifyDomain organization={organization} />}
                />
                <Route
                  path="modal"
                  element={<AddDomain organization={organization} />}
                />
              </Route>

              <Route
                path="teams/:teamId"
                element={<TeamPage organization={organization} />}
              >
                <Route
                  path="delete"
                  element={<DeleteTeam organization={organization} />}
                />
              </Route>

              <Route
                path="projects"
                element={<WorkspaceProjects organization={organization} />}
              >
                <Route
                  path="modal"
                  element={<AddProject organization={organization} />}
                />
              </Route>

              <Route
                path="projects/:projectId"
                element={<ProjectPage organization={organization} />}
              >
                <Route
                  path="delete"
                  element={<DeleteProject organization={organization} />}
                />
              </Route>

              <Route path="profile" element={<UserSetting />} />
              <Route path="perferences" element={<UserPreferences />} />
            </Routes>
          </Flex>
        </Flex>
      </ThemeProvider>
    </MemoryRouter>
  );
};
