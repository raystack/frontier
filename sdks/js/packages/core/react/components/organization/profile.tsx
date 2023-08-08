import { Flex, ThemeProvider } from '@raystack/apsara';
import { useEffect, useState } from 'react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { Toaster } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import GeneralSetting from './general';
import WorkspaceMembers from './members';
import { InviteMember } from './members/invite';
import UserPreferences from './preferences';
import WorkspaceSecurity from './security';
import { Sidebar } from './sidebar';
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
              />
              <Route path="/security" element={<WorkspaceSecurity />} />
              <Route
                path="/members"
                element={<WorkspaceMembers users={users}></WorkspaceMembers>}
              >
                <Route
                  path="modal"
                  element={
                    <InviteMember organization={organization} users={users} />
                  }
                />
              </Route>
              <Route path="/profile" element={<UserSetting />} />
              <Route path="/perferences" element={<UserPreferences />} />
            </Routes>
          </Flex>
        </Flex>
      </ThemeProvider>
    </MemoryRouter>
  );
};
