import React, { useContext } from "react";
import {
  Avatar,
  DropdownMenu,
  Flex,
  Sidebar,
  useTheme,
} from "@raystack/apsara/v1";
import { api } from "~/api";

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
import { MoonIcon, SunIcon } from "@radix-ui/react-icons";
import { useLocation } from "react-router-dom";

export type NavigationItemsTypes = {
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
  const location = useLocation();

  const isActive = (navlink?: string) => {
    const firstPathPart = location.pathname.split("/")[1];
    const firstPartOfNavlink = navlink?.split("/")[1];
    const isMatchingPath = firstPartOfNavlink === firstPathPart;
    return isMatchingPath;
  };
  return (
    <Sidebar open className={styles.sidebar}>
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
            >
              {nav.subItems?.map((subItem) => (
                <Sidebar.Item
                  icon={subItem.icon}
                  key={subItem.name}
                  active={isActive(subItem.to)}
                  href={subItem?.to}
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
              active={isActive(nav.to)}
              href={nav?.to}
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
    </Sidebar>
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

  const themeData =
    theme === "light"
      ? { icon: <MoonIcon />, label: "Dark" }
      : { icon: <SunIcon />, label: "Light" };

  return (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild>
        <Sidebar.Item
          icon={<Avatar src={user?.avatar} fallback={userInital} size={3} />}
          data-test-id="frontier-sdk-sidebar-logout"
        >
          {user?.email}
        </Sidebar.Item>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content>
        <DropdownMenu.Item onSelect={toggleTheme}>
          {themeData.icon} {themeData.label}
        </DropdownMenu.Item>
        <DropdownMenu.Item onSelect={logout}>Logout</DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu>
  );
}
