import React from "react";
import { ScrollArea, Sidebar } from "@raystack/apsara";

import { Flex, ThemeSwitcher } from "@raystack/apsara/v1";
import "@raystack/apsara/style.css";
import { Outlet, useNavigate } from "react-router-dom";
import "./App.css";
import { api } from "~/api";

export type NavigationItemsTypes = {
  active?: boolean;
  to?: string;
  name: string;
  icon?: React.ReactNode;
};

const BRAND_NAME = "Frontier";
const navigationItems: NavigationItemsTypes[] = [
  {
    name: "Organisations",
    to: `/organisations`,
  },
  {
    name: "Invoices",
    to: `/invoices`,
  },
  {
    name: "Projects",
    to: `/projects`,
  },
  {
    name: "Users",
    to: `/users`,
  },
  {
    name: "Groups",
    to: `/groups`,
  },
  {
    name: "Products",
    to: `/products`,
  },
  {
    name: "Plans",
    to: `/plans`,
  },
  {
    name: "Roles",
    to: `/roles`,
  },
  {
    name: "Preferences",
    to: `/preferences`,
  },
  {
    name: "Super Admins",
    to: `/super-admins`,
  },
  {
    name: "Webhooks",
    to: `/webhooks`,
  },
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
      <Sidebar>
        <Flex direction="column" style={{ maxWidth: "250px" }}>
          <Sidebar.Logo name={BRAND_NAME} />
          <Flex
            // @ts-ignore
            style={{
              marginTop: "22px",
              maxHeight: "calc(100vh - 120px)",
            }}
          >
            <ScrollArea style={{ paddingRight: "var(--mr-16)", width: "100%" }}>
              <Sidebar.Navigations>
                {navigationItems.map((nav) => (
                  <Sidebar.NavigationCell
                    key={nav.name}
                    active={nav.active}
                    onClick={() => navigate(nav?.to as string)}
                    data-test-id={`admin-ui-sidebar-navigation-cell-${nav.name}`}
                  >
                    {nav.name}
                  </Sidebar.NavigationCell>
                ))}
              </Sidebar.Navigations>
            </ScrollArea>
          </Flex>
        </Flex>
        <Sidebar.Footer
          action={
            // @ts-ignore
            <Sidebar.Navigations style={{ width: "100%" }}>
              <Sidebar.NavigationCell asChild>
                <ThemeSwitcher size={16} />
              </Sidebar.NavigationCell>
              <Sidebar.NavigationCell
                onClick={logout}
                data-test-id="frontier-sdk-sidebar-logout"
              >
                Logout
              </Sidebar.NavigationCell>
            </Sidebar.Navigations>
          }
        ></Sidebar.Footer>
      </Sidebar>
      <Flex style={{ flexGrow: "1", overflow: "auto" }}>
        <Outlet />
      </Flex>
    </div>
  );
}

export default App;
