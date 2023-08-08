import { Flex } from '@raystack/apsara';
import { useEffect, useState } from 'react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { Toaster } from 'sonner';
import { useFrontier } from '~/react/contexts/FrontierContext';
import GeneralSetting from './general';
import WorkspaceMembers from './members';
import WorkspaceSecurity from './security';
import { Sidebar } from './sidebar';
interface OrganizationProfileProps {
  organizationId: string;
}
const components = {
  general: <></>
};

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
              element={<WorkspaceMembers users={[]}></WorkspaceMembers>}
            />
          </Routes>
        </Flex>
      </Flex>
    </MemoryRouter>
  );
};;;;
