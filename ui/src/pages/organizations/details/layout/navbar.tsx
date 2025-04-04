import { SidebarIcon, OrganizationIcon } from "@raystack/apsara/icons";

import {
  Flex,
  Text,
  Breadcrumb,
  Avatar,
  IconButton,
  DropdownMenu,
  Chip,
  Spinner,
} from "@raystack/apsara/v1";

import styles from "./layout.module.css";
import { ChevronRightIcon, DotsHorizontalIcon } from "@radix-ui/react-icons";
import { V1Beta1Organization } from "~/api/frontier";
import { NavLink, useLocation } from "react-router-dom";
import { InviteUsersDialog } from "./invite-users-dialog";
import { AddTokensDialog } from "./add-tokens-dialog";

import React, { useContext, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { CollapsableSearch } from "~/components/collapsable-search";
import { api } from "~/api";

const downloadFile = (data: File, filename: string) => {
  const link = document.createElement("a");
  const downloadUrl = window.URL.createObjectURL(new Blob([data]));
  link.href = downloadUrl;
  link.setAttribute("download", filename);
  document.body.appendChild(link);
  link.click();
  link.parentNode?.removeChild(link);
  window.URL.revokeObjectURL(downloadUrl);
};

const NavbarActionMenu = ({ organizationId }: { organizationId: string }) => {
  const [isInviteUsersDialogOpen, setIsInviteUsersDialogOpen] = useState(false);
  const [isAddTokensDialogOpen, setIsAddTokensDialogOpen] = useState(false);
  const [isMembersDownloading, setIsMembersDownloading] = useState(false);

  const openInviteUsersDialog = () => {
    setIsInviteUsersDialogOpen(true);
  };

  const openAddTokensDialog = () => {
    setIsAddTokensDialogOpen(true);
  };

  async function handleExportMembers(e: Event) {
    e.preventDefault();
    try {
      setIsMembersDownloading(true);
      const response = await api.adminServiceExportOrganizationUsers(
        organizationId,
        {
          format: "blob",
        },
      );
      downloadFile(response.data, "members.csv");
    } catch (error) {
      console.error(error);
    } finally {
      setIsMembersDownloading(false);
    }
  }

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
      onSelect: openAddTokensDialog,
    },
    {
      label: "Change plan...",
      disabled: true,
    },
    {
      label: "Export members",
      onSelect: handleExportMembers,
      isLoading: isMembersDownloading,
    },
  ];

  return (
    <>
      {isInviteUsersDialogOpen ? (
        <InviteUsersDialog onOpenChange={setIsInviteUsersDialogOpen} />
      ) : null}
      {isAddTokensDialogOpen ? (
        <AddTokensDialog onOpenChange={setIsAddTokensDialogOpen} />
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
              className={styles["navbar-action-menu-item"]}
            >
              <Text>{item.label}</Text>
              {item.isLoading ? <Spinner size={2} /> : null}
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
}

export const OrganizationsDetailsNavabar = ({
  organization,
  toggleSidePanel,
}: OrganizationDetailsNavbarProps) => {
  const { search } = useContext(OrganizationContext);

  function handleSearchChange(e: React.ChangeEvent<HTMLInputElement>) {
    const value = e.target.value;
    if (value.length > 0 && search.onChange) {
      search.onChange(value);
    }
  }

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
        <NavbarActionMenu organizationId={organization.id || ""} />
        <NavLinks organizationId={organization.id || ""} />
      </Flex>
      <Flex align="center" gap={4}>
        {search.isVisible ? (
          <CollapsableSearch
            value={search.query}
            onChange={handleSearchChange}
            data-test-id="admin-ui-org-details-navbar-search"
          />
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
