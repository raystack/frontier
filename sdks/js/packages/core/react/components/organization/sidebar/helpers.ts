export type NavigationItemsTypes = {
  active?: boolean;
  to?: string;
  name: string;
  icon?: React.ReactNode;
};

export const organizationNavItems = [
  {
    name: 'General',
    to: '/'
  },
  {
    name: 'Security',
    to: '/security'
  },
  {
    name: 'Members',
    to: '/members'
  },
  {
    name: 'Teams',
    to: '/teams'
  },
  {
    name: 'Projects',
    to: '/projects'
  }
] as NavigationItemsTypes[];

export const userNavItems = [
  {
    name: 'Profile',
    to: '/profile'
  },
  {
    name: 'Perferences',
    to: '/perferences'
  }
] as NavigationItemsTypes[];
