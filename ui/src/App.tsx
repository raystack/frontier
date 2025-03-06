import React from "react";

import { Flex, Sidebar } from "@raystack/apsara/v1";
import "@raystack/apsara/style.css";
import { Outlet, useNavigate } from "react-router-dom";
import "./App.css";
import { api } from "~/api";
import { InfoCircledIcon, PersonIcon } from "@radix-ui/react-icons";

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
    name: "Organisations",
    to: `/organisations`,
  },
  {
    name: "Users",
    to: `/users`,
  },
  {
    name: "Invoices",
    to: `/invoices`,
  },
  {
    name: "Authorization",
    subItems: [
      {
        name: "Roles",
        to: `/roles`,
      },
    ],
  },
  {
    name: "Billing",
    subItems: [
      {
        name: "Products",
        to: `/products`,
      },
      {
        name: "Plans",
        to: `/plans`,
      },
    ],
  },
  {
    name: "Features",
    subItems: [
      {
        name: "Webhooks",
        to: `/webhooks`,
      },
    ],
  },
  {
    name: "Settings",
    subItems: [
      {
        name: "Preferences",
        to: `/preferences`,
      },
      {
        name: "Super Admins",
        to: `/super-admins`,
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

function App() {
  const navigate = useNavigate();

  async function logout() {
    await api?.frontierServiceAuthLogout();
    window.location.href = "/";
    window.location.reload();
  }
  return (
    <div style={{ display: "flex", height: "100vh", overflow: "hidden" }}>
      <Sidebar.Root open>
        <Sidebar.Header
          logo={
            <Flex align="center" style={{ height: "100%" }}>
              <InfoCircledIcon />
            </Flex>
          }
          title={BRAND_NAME}
        />
        <Sidebar.Main>
          {navigationItems.map((nav) => {
            return nav?.subItems?.length ? (
              <Sidebar.Group name={nav.name}>
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
          <Sidebar.Item
            icon={<span />}
            onClick={logout}
            data-test-id="frontier-sdk-sidebar-logout"
          >
            Logout
          </Sidebar.Item>
        </Sidebar.Footer>
      </Sidebar.Root>

      <Flex style={{ flexGrow: "1", overflow: "auto" }}>
        <Outlet />
      </Flex>
    </div>
  );
}

export default App;
