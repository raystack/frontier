import {
  Image,
  Sidebar as SidebarComponent,
  Flex,
  InputField
} from '@raystack/apsara/v1';
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
    <SidebarComponent open={true} className={styles.sidebarWrapper} disableResize>
      <div className={styles.scrollArea}>
        <Flex direction="column" gap={7} style={{ marginTop: '40px' }}>
          <InputField
            size="large"
            leadingIcon={
              <MagnifyingGlassIcon
                style={{ color: 'var(--rs-color-foreground-base-primary)' }}
              />
            }
            placeholder="Search"
            onChange={event => setSearch(event.target.value)}
            data-test-id="frontier-sdk-sidebar-search-field"
          />
          <SidebarComponent.Main>
            <SidebarComponent.Group
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
                    <SidebarComponent.Item
                      key={nav.name}
                      icon={<></>}
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
            </SidebarComponent.Group>
            <SidebarComponent.Group
              name="My Account"
              icon={<Image alt="user" width={16} height={16} src={user} />}
            >
              {userNavItems
                .filter(s => s.name.toLowerCase().includes(search))
                .map(nav => (
                  <SidebarComponent.Item
                    key={nav.name}
                    icon={<></>}
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
