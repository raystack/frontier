import React, { useContext } from "react";
import {
  Avatar,
  DropdownMenu,
  Flex,
  Sidebar,
  useTheme,
} from "@raystack/apsara/v1";
import { api } from "~/api";
import { useNavigate } from "react-router-dom";

import styles from "./sidebar.module.css";

import IAMIcon from "~/assets/icons/iam.svg?react";
import OrganizationsIcon from "~/assets/icons/organization.svg?react";
import UserIcon from "~/assets/icons/users.svg?react";
import InvoicesIcon from "~/assets/icons/invoices.svg?react";
import RolesIcon from "~/assets/icons/roles.svg?react";
import ProductsIcon from "~/assets/icons/products.svg?react";
import PlansIcon from "~/assets/icons/plans.svg?react";
import WebhooksIcon from "~/assets/icons/webhooks.svg?react";
import PreferencesIcon from "~/assets/icons/preferences.svg?react";
import AdminsIcon from "~/assets/icons/admins.svg?react";
import { AppContext } from "~/contexts/App";

export type NavigationItemsTypes = {
  active?: boolean;
  to?: string;
  name: string;
  icon?: React.ReactNode;
  subItems?: NavigationItemsTypes[];
};

const BRAND_NAME = "Frontier";

const navigationItems: NavigationItemsTypes[] = [
  {
    name: "Organizations",
    to: `/organisations`,
    icon: <OrganizationsIcon />,
  },
  {
    name: "Users",
    to: `/users`,
    icon: <UserIcon />,
  },
  {
    name: "Invoices",
    to: `/invoices`,
    icon: <InvoicesIcon />,
  },
  {
    name: "Authorization",
    subItems: [
      {
        name: "Roles",
        to: `/roles`,
        icon: <RolesIcon />,
      },
    ],
  },
  {
    name: "Billing",
    subItems: [
      {
        name: "Products",
        to: `/products`,
        icon: <ProductsIcon />,
      },
      {
        name: "Plans",
        to: `/plans`,
        icon: <PlansIcon />,
      },
    ],
  },
  {
    name: "Features",
    subItems: [
      {
        name: "Webhooks",
        to: `/webhooks`,
        icon: <WebhooksIcon />,
      },
    ],
  },
  {
    name: "Settings",
    subItems: [
      {
        name: "Preferences",
        to: `/preferences`,
        icon: <PreferencesIcon />,
      },
      {
        name: "Admins",
        to: `/super-admins`,
        icon: <AdminsIcon />,
      },
    ],
  },
  // {
  //   name: "Projects",
  //   to: `/projects`,
  // },

  // {
  //   name: "Groups",
  //   to: `/groups`,
  // },
];

export default function IAMSidebar() {
  const navigate = useNavigate();

  return (
    <Sidebar.Root open className={styles.sidebar}>
      <Sidebar.Header
        logo={
          <Flex align="center" style={{ height: "100%" }}>
            <IAMIcon />
          </Flex>
        }
        title={BRAND_NAME}
      />
      <Sidebar.Main>
        {navigationItems.map((nav) => {
          return nav?.subItems?.length ? (
            <Sidebar.Group
              name={nav.name}
              key={nav.name}
              className={styles["sidebar-group"]}
              // TODO: pass className to the components
              style={{ marginTop: "var(--rs-space-5)" }}
            >
              {nav.subItems?.map((subItem) => (
                <Sidebar.Item
                  icon={subItem.icon}
                  key={subItem.name}
                  active={subItem.active}
                  onClick={() => navigate(subItem?.to as string)}
                  data-test-id={`admin-ui-sidebar-navigation-cell-${subItem.name}`}
                >
                  {subItem.name}
                </Sidebar.Item>
              ))}
            </Sidebar.Group>
          ) : (
            <Sidebar.Item
              icon={nav.icon}
              key={nav.name}
              active={nav.active}
              onClick={() => navigate(nav?.to as string)}
              data-test-id={`admin-ui-sidebar-navigation-cell-${nav.name}`}
            >
              {nav.name}
            </Sidebar.Item>
          );
        })}
      </Sidebar.Main>
      <Sidebar.Footer>
        <UserDropdown />
      </Sidebar.Footer>
    </Sidebar.Root>
  );
}

function UserDropdown() {
  const { user } = useContext(AppContext);
  const { theme, setTheme } = useTheme();

  async function logout() {
    await api?.frontierServiceAuthLogout();
    window.location.href = "/";
    window.location.reload();
  }

  const toggleTheme = () => {
    if (theme === "dark") {
      setTheme("light");
    } else {
      setTheme("dark");
    }
  };

  const userInital = user?.title?.[0] || user?.email?.[0];

  return (
    <DropdownMenu>
      <DropdownMenu.Trigger>
        <Sidebar.Item
          icon={<Avatar src={user?.avatar} fallback={userInital} size={2} />}
          data-test-id="frontier-sdk-sidebar-logout"
        >
          {user?.email}
        </Sidebar.Item>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content>
        <DropdownMenu.Item onSelect={toggleTheme}>{theme}</DropdownMenu.Item>
        <DropdownMenu.Item onSelect={logout}>Logout</DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu>
  );
}
