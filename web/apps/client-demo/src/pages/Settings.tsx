import { useEffect } from 'react';
import { Flex, Sidebar, Text } from '@raystack/apsara';
import { Outlet, useParams, useLocation, Navigate } from 'react-router-dom';
import { useFrontier } from '@raystack/frontier/react';

const NAV_ITEMS = [
  { label: 'General', path: 'general' },
  { label: 'Preferences', path: 'preferences' }
];

export default function Settings() {
  const { orgId } = useParams<{ orgId: string }>();
  const location = useLocation();
  const { organizations, setActiveOrganization, activeOrganization } =
    useFrontier();

  useEffect(() => {
    if (!orgId || organizations.length === 0) return;
    const org = organizations.find(_org => _org.id === orgId || _org.name === orgId);
    if (org && activeOrganization?.id !== org.id) {
      setActiveOrganization(org);
    }
  }, [orgId, organizations, activeOrganization?.id, setActiveOrganization]);

  if (!orgId) return null;

  const isSettingsRoot = location.pathname === `/${orgId}/settings`;
  if (isSettingsRoot) {
    return <Navigate to={`/${orgId}/settings/general`} replace />;
  }

  return (
    <Flex style={{ height: '100vh', width: '100vw' }}>
      <Sidebar defaultOpen>
        <Sidebar.Header>
          <Flex align="center" gap={3}>
            <Text size={4} weight="medium" data-collapse-hidden>
              Settings
            </Text>
          </Flex>
        </Sidebar.Header>
        <Sidebar.Main>
          <Sidebar.Group label="Organization">
            {NAV_ITEMS.map(item => {
              const fullPath = `/${orgId}/settings/${item.path}`;
              const isActive = location.pathname === fullPath;
              return (
                <Sidebar.Item
                  key={item.path}
                  href={fullPath}
                  active={isActive}
                  data-test-id={`[settings-nav-${item.path}]`}
                >
                  {item.label}
                </Sidebar.Item>
              );
            })}
          </Sidebar.Group>
        </Sidebar.Main>
      </Sidebar>
      <Flex style={{ flex: 1, overflow: 'auto' }}>
        <Outlet />
      </Flex>
    </Flex>
  );
}
