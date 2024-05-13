import React from 'react';

export type NavigationItemsTypes = {
  active?: boolean;
  to?: string;
  name: string;
  icon?: React.ReactNode;
};

interface getOrganizationNavItemsOptions {
  tempShowBilling?: boolean;
  tempShowTokens?: boolean;
  canSeeBilling?: boolean;
}

export const getOrganizationNavItems = (
  options: getOrganizationNavItemsOptions = {}
) =>
  [
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
      show: options?.tempShowBilling && options?.canSeeBilling
    },
    {
      name: 'Tokens',
      to: '/tokens',
      show: options?.tempShowTokens
    },
    {
      name: 'Plans',
      to: '/plans',
      show: options?.tempShowBilling
    }
  ].filter(nav => nav.show) as NavigationItemsTypes[];

export const userNavItems = [
  {
    name: 'Profile',
    to: '/profile'
  }
  // {
  //   name: 'Preferences',
  //   to: '/preferences'
  // }
] as NavigationItemsTypes[];
