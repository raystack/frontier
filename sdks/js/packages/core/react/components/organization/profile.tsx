import { Flex, ThemeProvider } from '@raystack/apsara';
import { useEffect, useState } from 'react';
import {
  MemoryRouter,
  Route,
  Routes,
  UNSAFE_LocationContext
} from 'react-router-dom';
import { Toaster } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import Domain from './domain';
import { AddDomain } from './domain/add-domain';
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
}

export const OrganizationProfile = ({
  organizationId
}: OrganizationProfileProps) => {
  const [organization, setOrganization] = useState();
  const [users, setUsers] = useState([]);
  const { client } = useFrontier();

  useEffect(() => {
    async function fetchDetails() {
      const {
        // @ts-ignore
        data: { organization }
      } = await client?.frontierServiceGetOrganization(organizationId);
      setOrganization(organization);

      const {
        // @ts-ignore
        data: { users }
      } = await client?.frontierServiceListOrganizationUsers(organizationId);
      setUsers(users);
    }

    if (organizationId) fetchDetails();
  }, [organizationId, client]);

  return (
    // @ts-ignore
    <UNSAFE_LocationContext.Provider value={null}>
      <MemoryRouter>
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
                  element={<WorkspaceMembers users={users}></WorkspaceMembers>}
                >
                  <Route
                    path="modal"
                    element={
                      <InviteMember organization={organization} users={users} />
                    }
                  />
                  <Route
                    path="domain"
                    element={<AddDomain organization={organization} />}
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
    </UNSAFE_LocationContext.Provider>
  );
};
