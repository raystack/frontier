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
import { useState } from "react";

const NavbarActionMenu = () => {
  const [isInviteUsersDialogOpen, setIsInviteUsersDialogOpen] = useState(false);

  const openInviteUsersDialog = () => {
    setIsInviteUsersDialogOpen(true);
  };

  const items = [
    {
      label: "Edit...",
      disabled: true,
    },
    {
      label: "Invite users...",
      onSelect: openInviteUsersDialog,
    },
    {
      label: "Add project...",
      disabled: true,
    },
    {
      label: "Add tokens...",
      disabled: true,
    },
    {
      label: "Change plan...",
      disabled: true,
    },
  ];

  return (
    <>
      {isInviteUsersDialogOpen ? (
        <InviteUsersDialog onOpenChange={setIsInviteUsersDialogOpen} />
      ) : null}
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
          {items.map((item, index) => (
            <DropdownMenu.Item
              key={index}
              disabled={item.disabled}
              onSelect={item?.onSelect}
            >
              <Text>{item.label}</Text>
            </DropdownMenu.Item>
          ))}
        </DropdownMenu.Content>
      </DropdownMenu>
    </>
  );
};

const NavLinks = ({ organizationId }: { organizationId: string }) => {
  const location = useLocation();
  const currentPath = location.pathname;

  const links = [
    { name: "Members", path: `/organisations/${organizationId}/members` },
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
  isSearchVisible: boolean;
}

export const OrganizationsDetailsNavabar = ({
  organization,
  toggleSidePanel,
  isSearchVisible = false,
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
        {isSearchVisible ? (
          <IconButton size={3} data-test-id="admin-ui-nav-search-button">
            <MagnifyingGlassIcon />
          </IconButton>
        ) : null}
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
