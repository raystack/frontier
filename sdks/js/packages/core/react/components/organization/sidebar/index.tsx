import {
  Flex,
  ScrollArea,
  Sidebar as SidebarComponent,
  Text,
  TextField
} from '@raystack/apsara';
import { Image } from '@raystack/apsara/v1';
import { Link, useRouteContext, useRouterState } from '@tanstack/react-router';
import React, { useCallback, useMemo, useState } from 'react';
import organization from '~/react/assets/organization.png';
import user from '~/react/assets/user.png';
import { getOrganizationNavItems, getUserNavItems } from './helpers';

// @ts-ignore
import { MagnifyingGlassIcon } from '@radix-ui/react-icons';
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
    <SidebarComponent>
      <ScrollArea className={styles.scrollarea}>
        <Flex direction="column" style={{ gap: '24px', marginTop: '40px' }}>
          <TextField
            // @ts-ignore
            size="medium"
            // @ts-ignore
            leading={
              <MagnifyingGlassIcon
                style={{ color: 'var(--rs-color-foreground-base-primary)' }}
              />
            }
            placeholder="Search"
            onChange={event => setSearch(event.target.value)}
            data-test-id="frontier-sdk-sidebar-search-field"
          />
          <SidebarComponent.Navigations>
            <SidebarComponent.NavigationGroup
              name="Organization"
              icon={
                <Image
                  alt="organization"
                  width={16}
                  height={16}
                  src={organization}
                />
              }
            >
              {organizationNavItems
                .filter(s => s.name.toLowerCase().includes(search))
                .map(nav => {
                  return (
                    <SidebarComponent.NavigationCell
                      key={nav.name}
                      asChild
                      active={!!isActive(nav?.to as string) as any}
                      style={{ padding: 0 }}
                    >
                      <Link
                        key={nav.name}
                        to={nav.to as string}
                        data-test-id={`frontier-sdk-sidebar-link-${nav.name}`}
                        style={{
                          width: '100%',
                          textDecoration: 'none',
                          padding: 'var(--pd-8)'
                        }}
                        search={{}}
                        params={{}}
                      >
                        <Text
                          style={{
                            color: 'var(--rs-color-foreground-base-primary)',
                            fontWeight: 'var(--rs-font-weight-medium)'
                          }}
                        >
                          {nav.name}
                        </Text>
                      </Link>
                    </SidebarComponent.NavigationCell>
                  );
                })}
            </SidebarComponent.NavigationGroup>
            <SidebarComponent.NavigationGroup
              name="My Account"
              icon={<Image alt="user" width={16} height={16} src={user} />}
            >
              {userNavItems
                .filter(s => s.name.toLowerCase().includes(search))
                .map(nav => (
                  <SidebarComponent.NavigationCell
                    key={nav.name}
                    asChild
                    active={!!isActive(nav?.to as string) as any}
                    style={{ padding: 0 }}
                  >
                    <Link
                      key={nav.name}
                      to={nav.to as string}
                      data-test-id={`frontier-sdk-sidebar-link-${nav.name}`}
                      style={{
                        width: '100%',
                        textDecoration: 'none',
                        padding: 'var(--pd-8)'
                      }}
                      search={{}}
                      params={{}}
                    >
                      <Text
                        style={{
                          color: 'var(--rs-color-foreground-base-primary)',
                          fontWeight: 'var(--rs-font-weight-medium)'
                        }}
                      >
                        {nav.name}
                      </Text>
                    </Link>
                  </SidebarComponent.NavigationCell>
                ))}
            </SidebarComponent.NavigationGroup>
          </SidebarComponent.Navigations>
        </Flex>
      </ScrollArea>
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
