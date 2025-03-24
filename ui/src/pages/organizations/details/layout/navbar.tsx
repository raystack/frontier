import { SidebarIcon, OrganizationIcon } from "@raystack/apsara/icons";

import {
  Flex,
  Text,
  Breadcrumb,
  Avatar,
  IconButton,
  DropdownMenu,
  Chip,
} from "@raystack/apsara/v1";

import styles from "./layout.module.css";
import { ChevronRightIcon, DotsHorizontalIcon } from "@radix-ui/react-icons";
import { V1Beta1Organization } from "~/api/frontier";
import { NavLink, useLocation } from "react-router-dom";
import { InviteUsersDialog } from "./invite-users-dialog";

const NavbarActionMenu = () => {
  const preventDefault = (e: Event) => {
    e.preventDefault();
  };

  return (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild>
        <IconButton size={2} data-test-id="admin-ui-nav-action-menu-button">
          <DotsHorizontalIcon />
        </IconButton>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content
        className={styles["navbar-action-menu-content"]}
        align="start"
      >
        <DropdownMenu.Item>
          <Text>Edit...</Text>
        </DropdownMenu.Item>
        <InviteUsersDialog>
          <DropdownMenu.Item onSelect={preventDefault}>
            <Text>Invite users...</Text>
          </DropdownMenu.Item>
        </InviteUsersDialog>

        <DropdownMenu.Item>
          <Text>Add project...</Text>
        </DropdownMenu.Item>
        <DropdownMenu.Item>
          <Text>Add tokens...</Text>
        </DropdownMenu.Item>
        <DropdownMenu.Item>
          <Text>Change plan...</Text>
        </DropdownMenu.Item>
        {/* TODO: use submenus for exports */}
        {/* <DropdownMenu.Group>
          <DropdownMenu.Item>
            <Text>Members</Text>
          </DropdownMenu.Item>
          <DropdownMenu.Item>
            <Text>Projects</Text>
          </DropdownMenu.Item>
          <DropdownMenu.Item>
            <Text>Tokens</Text>
          </DropdownMenu.Item>
          <DropdownMenu.Item>
            <Text>Audit log</Text>
          </DropdownMenu.Item>
        </DropdownMenu.Group> */}
      </DropdownMenu.Content>
    </DropdownMenu>
  );
};

const NavLinks = ({ organizationId }: { organizationId: string }) => {
  const location = useLocation();
  const currentPath = location.pathname;

  const links = [
    { name: "Members", path: `/organisations/${organizationId}/#` },
    { name: "Projects", path: `/organisations/${organizationId}/#` },
    { name: "Tokens", path: `/organisations/${organizationId}/#` },
    { name: "API", path: `/organisations/${organizationId}/#` },
    { name: "Audit log", path: `/organisations/${organizationId}/#` },
    { name: "Security", path: `/organisations/${organizationId}/security` },
  ];

  function checkActive(path: string) {
    return currentPath.startsWith(path);
  }

  return (
    <Flex gap={3}>
      {links.map((link, i) => {
        const isActive = checkActive(link.path);
        return (
          <NavLink to={link.path} key={link.path + i}>
            <Chip data-state={isActive ? "active" : ""} variant={"filled"}>
              {link.name}
            </Chip>
          </NavLink>
        );
      })}
    </Flex>
  );
};

interface OrganizationDetailsNavbarProps {
  organization: V1Beta1Organization;
  toggleSidePanel: () => void;
}

export const OrganizationsDetailsNavabar = ({
  organization,
  toggleSidePanel,
}: OrganizationDetailsNavbarProps) => {
  return (
    <nav className={styles.navbar}>
      <Flex gap={4} align="center">
        <Breadcrumb
          size="small"
          separator={<ChevronRightIcon style={{ display: "flex" }} />}
          items={[
            {
              label: "Organizations",
              href: "/organisations",
              icon: <OrganizationIcon />,
            },
            {
              label: organization?.title || "NA",
              href: `/organigations/${organization?.id}`,
              icon: (
                <Avatar
                  src={organization?.avatar}
                  fallback={organization?.title?.[0]}
                  size={1}
                />
              ),
            },
          ]}
        />
        <NavbarActionMenu />
        <NavLinks organizationId={organization.id || ""} />
      </Flex>
      <Flex align="center" gap={4}>
        {/* <IconButton size={3} data-test-id="admin-ui-nav-search-button">
          <MagnifyingGlassIcon />
        </IconButton> */}
        <IconButton
          size={3}
          data-test-id="admin-ui-nav-sidepanel-button"
          onClick={toggleSidePanel}
        >
          <SidebarIcon />
        </IconButton>
      </Flex>
    </nav>
  );
};
