import { Box, styled, Text } from "@odpf/apsara";
import { NavLink } from "react-router-dom";

export type NavigationItemsTypes = {
  active?: boolean;
  to?: string;
  name: string;
  icon?: React.ReactNode;
};

const NavItem = styled(NavLink, {
  cursor: "pointer",
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  padding: "8px",
  color: "$gray12",
  textDecoration: "none",
  marginBottom: "4px",
  svg: {
    width: "1.8rem",
    height: "1.6rem",
  },
  "& + a": {},
  "&:hover,&:focus,&:active, &.active": {
    borderRadius: "$1",
    backgroundColor: "$gray4",
  },
});

export default function SidebarContent() {
  const navigationItems: NavigationItemsTypes[] = [
    {
      name: "Dashboard",
      icon: "dashboard",
      to: `/console/dashboard`,
    },
    {
      name: "Organisations",
      icon: "organisations",
      to: `/console/organisations`,
    },
    {
      name: "Projects",
      icon: "projects",
      to: `/console/projects`,
    },
    {
      name: "Users",
      icon: "users",
      to: `/console/users`,
    },
    {
      name: "Groups",
      icon: "groups",
      to: `/console/groups`,
    },
    {
      name: "Roles",
      icon: "groups",
      to: `/console/roles`,
    },
  ];
  return (
    <Box css={sidebarNavigationContainerStyle}>
      {navigationItems.map(({ active, name, icon, to }) => (
        <NavItem key={name} className={`${active && "active"} `} to={to || ""}>
          <img src={`/console/${icon}.svg`} />
          <Text size={2} css={{ fontWeight: "500", marginLeft: "8px" }}>
            {name}
          </Text>
        </NavItem>
      ))}
    </Box>
  );
}
const sidebarNavigationContainerStyle = {
  margin: "1rem",
};
