import { Link } from "react-router-dom";
import { Flex } from "~/primitives/flex/Flex";
import { Text } from "~/primitives/text/Text";
import { styled } from "~/stitches";

export type NavigationItemsTypes = {
  active?: boolean;
  to?: string;
  name: string;
  icon?: React.ReactNode;
};

const NavItem = styled(Link, {
  cursor: "pointer",
  height: "3.2rem",
  display: "flex",
  flexDirection: "row",
  alignItems: "center",
  padding: "$8",
  color: "$gray12",
  textDecoration: "none",
  marginBottom: "$4",
  svg: {
    width: "1.8rem",
    height: "1.6rem",
  },
  "& + a": {},
  "&:hover,&:focus,&:active, &.active": {
    borderRadius: "$4",
    backgroundColor: "$gray4",
  },
});

export default function SidebarContent() {
  const navigationItems: NavigationItemsTypes[] = [
    {
      active: true,
      name: "Home",
      to: `/`,
    },
    {
      active: true,
      name: "Dashboard",
      to: `/dashboard`,
    },
  ];
  return (
    <Flex as="nav" css={sidebarNavigationContainerStyle}>
      {navigationItems.map(({ active, name, icon, to }) => (
        <NavItem key={name} className={`${active && "active"} `} to={to || ""}>
          {icon}
          <Text css={{ marginLeft: "$8", fontWeight: "500" }}>{name}</Text>
        </NavItem>
      ))}
    </Flex>
  );
}
const sidebarNavigationContainerStyle = {};
