import {
  Flex,
  Image,
  ScrollArea,
  Sidebar as SidebarComponent,
  Text,
  TextField
} from '@raystack/apsara';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import organization from '~/react/assets/organization.png';
import user from '~/react/assets/user.png';
import { organizationNavItems, userNavItems } from './helpers';

export const Sidebar = () => {
  const [search, setSearch] = useState('');
  return (
    <SidebarComponent>
      <ScrollArea
        style={{
          paddingRight: 'var(--mr-16)',
          width: '100%'
        }}
      >
        <Flex direction="column" style={{ gap: '24px' }}>
          <SidebarComponent.Logo name="" />
          <TextField
            // @ts-ignore
            size="medium"
            placeholder="Search"
            onChange={event => setSearch(event.target.value)}
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
                .map(nav => (
                  <SidebarComponent.NavigationCell
                    key={nav.name}
                    asChild
                    style={{ padding: 0 }}
                  >
                    <Link
                      key={nav.name}
                      to={nav.to as string}
                      style={{
                        width: '100%',
                        textDecoration: 'none',
                        padding: 'var(--pd-8)'
                      }}
                    >
                      <Text
                        style={{
                          color: 'var(--foreground-base)',
                          fontWeight: '500'
                        }}
                      >
                        {nav.name}
                      </Text>
                    </Link>
                  </SidebarComponent.NavigationCell>
                ))}
            </SidebarComponent.NavigationGroup>
            <SidebarComponent.NavigationGroup
              name="My Account"
              icon={<Image alt="user" width={16} height={16} src={user} />}
            >
              {userNavItems
                .filter(s => s.name.toLowerCase().includes(search))
                .map(nav => (
                  <SidebarComponent.NavigationCell key={nav.name} asChild>
                    <Link
                      key={nav.name}
                      to={nav.to as string}
                      style={{
                        textDecoration: 'none'
                      }}
                    >
                      <Text
                        style={{
                          color: 'var(--foreground-base)',
                          fontWeight: '500'
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
