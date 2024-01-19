import { Flex, ScrollArea, Sidebar } from "@raystack/apsara";
import "@raystack/apsara/index.css";
import { Outlet, useNavigate } from "react-router-dom";
import "./App.css";

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
    name: "Roles",
    to: `/roles`,
  },
];

function App() {
  const navigate = useNavigate();
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
                  >
                    {nav.name}
                  </Sidebar.NavigationCell>
                ))}
              </Sidebar.Navigations>
            </ScrollArea>
          </Flex>
        </Flex>
        <Sidebar.Footer></Sidebar.Footer>
      </Sidebar>
      <Flex style={{ flexGrow: "1", overflow: "auto" }}>
        <Outlet />
      </Flex>
    </div>
  );
}

export default App;
