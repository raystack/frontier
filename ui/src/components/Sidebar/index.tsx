import type React from "react";
import { useContext } from "react";
import {
  Avatar,
  DropdownMenu,
  Flex,
  Sidebar,
  Text,
  useTheme,
} from "@raystack/apsara/v1";
import { useMutation } from "@connectrpc/connect-query";
import { FrontierServiceQueries } from "@raystack/proton/frontier";

import styles from "./sidebar.module.css";
import { OrganizationIcon } from "@raystack/apsara/icons";
import IAMIcon from "~/assets/icons/iam.svg?react";
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
import { Link, useLocation } from "react-router-dom";

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
    to: `/organizations`,
    icon: <OrganizationIcon />,
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
    <Sidebar open className={styles.sidebar} collapsible={false}>
      <Sidebar.Header>
        <Flex align="center" style={{ height: "100%" }}>
          <IAMIcon />
        </Flex>
        <Text size="small" weight="medium">
          {BRAND_NAME}
        </Text>
      </Sidebar.Header>
      <Sidebar.Main>
        {navigationItems.map(nav => {
          return nav?.subItems?.length ? (
            <Sidebar.Group
              label={nav.name}
              key={nav.name}
              className={styles["sidebar-group"]}>
              {nav.subItems?.map(subItem => (
                <Sidebar.Item
                  leadingIcon={subItem.icon}
                  key={subItem.name}
                  active={isActive(subItem.to)}
                  data-test-id={`admin-ui-sidebar-navigation-cell-${subItem.name}`}
                  as={<Link to={subItem?.to ?? ""} />}>
                  {subItem.name}
                </Sidebar.Item>
              ))}
            </Sidebar.Group>
          ) : (
            <Sidebar.Item
              leadingIcon={nav.icon}
              key={nav.name}
              active={isActive(nav.to)}
              data-test-id={`admin-ui-sidebar-navigation-cell-${nav.name}`}
              as={<Link to={nav?.to ?? ""} />}>
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
  const logoutMutation = useMutation(FrontierServiceQueries.authLogout, {
    onSuccess: () => {
      window.location.href = "/";
      window.location.reload();
    },
  });

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
    <DropdownMenu placement="top">
      <DropdownMenu.Trigger asChild>
        <Sidebar.Item
          leadingIcon={
            <Avatar src={user?.avatar} fallback={userInital} size={3} />
          }
          data-test-id="frontier-sdk-sidebar-logout">
          {user?.email}
        </Sidebar.Item>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content>
        <DropdownMenu.Item
          onClick={toggleTheme}
          data-test-id="admin-ui-toggle-theme">
          {themeData.icon} {themeData.label}
        </DropdownMenu.Item>
        <DropdownMenu.Item onClick={() => logoutMutation.mutate({})} data-test-id="admin-ui-logout-btn">
          Logout
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu>
  );
}
