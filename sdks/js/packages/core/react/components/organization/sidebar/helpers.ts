import React from 'react';

export type NavigationItemsTypes = {
  active?: boolean;
  to?: string;
  name: string;
  icon?: React.ReactNode;
};

interface getOrganizationNavItemsOptions {
  showBilling?: boolean;
  showTokens?: boolean;
  canSeeBilling?: boolean;
}

interface getUserNavItemsOptions {
  showPreferences?: boolean;
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
    }
  ].filter(nav => nav.show) as NavigationItemsTypes[];

export const getUserNavItems = (options: getUserNavItemsOptions = {}) =>
  [
    {
      name: 'Profile',
      to: '/profile',
      show: true
    },
    {
      name: 'Preferences',
      to: '/preferences',
      show: options?.showPreferences
    }
  ].filter(nav => nav.show) as NavigationItemsTypes[];
