import {
  Image,
  Sidebar as SidebarComponent,
  Flex,
  Search
} from '@raystack/apsara/v1';
import { Link, useRouteContext, useRouterState } from '@tanstack/react-router';
import React, { useCallback, useMemo, useState } from 'react';
import organization from '~/react/assets/organization.png';
import user from '~/react/assets/user.png';
import { getOrganizationNavItems, getUserNavItems } from './helpers';

import { usePermissions } from '~/react/hooks/usePermissions';
import { PERMISSIONS, shouldShowComponent } from '~/utils';
import styles from './sidebar.module.css';

export const Sidebar = () => {
  const [search, setSearch] = useState('');
  const routerState = useRouterState();
  const {
    organizationId,
    showBilling,
    showTokens,
    showAPIKeys,
    showPreferences,
    customRoutes
  } = useRouteContext({
    from: '__root__'
  });

  const isActive = useCallback(
    (path: string) =>
      path.length > 2
        ? routerState.location.pathname.includes(path)
        : routerState.location.pathname === path,
    [routerState.location.pathname]
  );

  const resource = `app/organization:${organizationId}`;
  const listOfPermissionsToCheck = useMemo(
    () => [
      {
        permission: PERMISSIONS.UpdatePermission,
        resource
      }
    ],
    [resource]
  );

  const { permissions, isFetching: isPermissionsFetching } = usePermissions(
    listOfPermissionsToCheck,
    !!organizationId
  );

  const { canSeeBilling } = useMemo(() => {
    return {
      canSeeBilling: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
      )
    };
  }, [permissions, resource]);

  const organizationNavItems = useMemo(
    () =>
      getOrganizationNavItems({
        showBilling: showBilling,
        canSeeBilling: canSeeBilling,
        showTokens: showTokens,
        showAPIKeys: showAPIKeys,
        customRoutes: customRoutes.Organization
      }),
    [
      showBilling,
      canSeeBilling,
      showTokens,
      showAPIKeys,
      customRoutes.Organization
    ]
  );

  const userNavItems = useMemo(
    () =>
      getUserNavItems({
        showPreferences: showPreferences,
        customRoutes: customRoutes.User
      }),
    [customRoutes.User, showPreferences]
  );

  return (
    <SidebarComponent
      open
      className={styles.sidebarWrapper}
      collapsible={false}
    >
      <div className={styles.scrollArea}>
        <Flex direction="column" gap={4} style={{ marginTop: '64px' }}>
          <Search
            size="large"
            value={search}
            onChange={event => setSearch(event.target.value)}
            placeholder="Search pages"
            showClearButton
            onClear={() => setSearch('')}
            data-test-id="frontier-sdk-sidebar-search-field"
          />

          <SidebarComponent.Main style={{ gap: 'var(--rs-space-5)' }}>
            <SidebarComponent.Group
              label="Organization"
              leadingIcon={
                <Image
                  alt="organization"
                  width={16}
                  height={16}
                  src={organization}
                />
              }
              className={styles.sidebarItemGroupContainer}
            >
              <Flex
                direction="column"
                gap={2}
                className={styles.sidebarItemGroup}
              >
                {organizationNavItems
                  .filter(s => s.name.toLowerCase().includes(search))
                  .map(nav => {
                    return (
                      <SidebarComponent.Item
                        key={nav.name}
                        classNames={{
                          leadingIcon: styles.sidebarItemIcon
                        }}
                        as={
                          <Link
                            to={nav.to as string}
                            data-test-id={`frontier-sdk-sidebar-link-${nav.name}`}
                            style={{
                              width: '100%',
                              textDecoration: 'none'
                            }}
                            search={{}}
                            params={{}}
                          />
                        }
                        active={!!isActive(nav?.to as string)}
                      >
                        {nav.name}
                      </SidebarComponent.Item>
                    );
                  })}
              </Flex>
            </SidebarComponent.Group>
            <SidebarComponent.Group
              label="My Account"
              leadingIcon={
                <Image alt="user" width={16} height={16} src={user} />
              }
              className={styles.sidebarItemGroupContainer}
            >
              <Flex
                direction="column"
                gap={2}
                className={styles.sidebarItemGroup}
              >
                {userNavItems
                  .filter(s => s.name.toLowerCase().includes(search))
                  .map(nav => (
                    <SidebarComponent.Item
                      key={nav.name}
                      classNames={{
                        leadingIcon: styles.sidebarItemIcon
                      }}
                      as={
                        <Link
                          to={nav.to as string}
                          data-test-id={`frontier-sdk-sidebar-link-${nav.name}`}
                          style={{
                            width: '100%',
                            textDecoration: 'none'
                          }}
                          search={{}}
                          params={{}}
                        />
                      }
                      active={!!isActive(nav?.to as string)}
                    >
                      {nav.name}
                    </SidebarComponent.Item>
                  ))}
              </Flex>
            </SidebarComponent.Group>
          </SidebarComponent.Main>
        </Flex>
      </div>
    </SidebarComponent>
  );
};

type SidebarHeaderProps = { children?: React.ReactNode };
export const SidebarHeader = ({ children }: SidebarHeaderProps) => {
  return <Flex justify="between">{children}</Flex>;
};

type SidebarFooterProps = { children?: React.ReactNode };
export const SidebarFooter = ({ children }: SidebarFooterProps) => {
  return <Flex>{children}</Flex>;
};
