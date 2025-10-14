import React from 'react';

export type NavigationItemsTypes = {
  active?: boolean;
  to?: string;
  name: string;
  icon?: React.ReactNode;
};

type CustomRoutes = Array<{ name: string; path: string }>;

interface getOrganizationNavItemsOptions {
  showBilling?: boolean;
  showTokens?: boolean;
  showAPIKeys?: boolean;
  canSeeBilling?: boolean;
  customRoutes?: CustomRoutes;
}

interface getUserNavItemsOptions {
  showPreferences?: boolean;
  customRoutes?: CustomRoutes;
}

function getCustomRoutes(customRoutes: CustomRoutes = []) {
  return (
    customRoutes?.map(r => ({
      name: r.name,
      to: r.path,
      show: true
    })) || []
  );
}

export const getOrganizationNavItems = (
  options: getOrganizationNavItemsOptions = {}
) => {
  const routes = [
    {
      name: 'General',
      to: '/',
      show: true
    },
    {
      name: 'Members',
      to: '/members',
      show: true
    },
    {
      name: 'Teams',
      to: '/teams',
      show: true
    },
    {
      name: 'Projects',
      to: '/projects',
      show: true
    },
    {
      name: 'Security',
      to: '/domains',
      show: true
    },
    {
      name: 'Billing',
      to: '/billing',
      show: options?.showBilling && options?.canSeeBilling
    },
    {
      name: 'Tokens',
      to: '/tokens',
      show: options?.showTokens
    },
    {
      name: 'Plans',
      to: '/plans',
      show: options?.showBilling
    },
    {
      name: 'API',
      to: '/api-keys',
      show: options?.showAPIKeys
    }
  ];
  const customRoutes = getCustomRoutes(options?.customRoutes);
  return [...routes, ...customRoutes].filter(
    nav => nav.show
  ) as NavigationItemsTypes[];
};

export const getUserNavItems = (options: getUserNavItemsOptions = {}) => {
  const routes = [
    {
      name: 'Profile',
      to: '/profile',
      show: true
    },
    {
      name: 'Preferences',
      to: '/preferences',
      show: options?.showPreferences
    },
    {
      name: 'Sessions',
      to: '/sessions',
      show: true
    },
  ];
  const customRoutes = getCustomRoutes(options?.customRoutes);
  return [...routes, ...customRoutes].filter(
    nav => nav.show
  ) as NavigationItemsTypes[];
};
