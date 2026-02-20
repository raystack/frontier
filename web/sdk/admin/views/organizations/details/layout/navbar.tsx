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
  getAvatarColor,
  toast,
} from "@raystack/apsara";

import styles from "./layout.module.css";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import { NavLink, useLocation } from "react-router-dom";
import { InviteUsersDialog } from "./invite-users-dialog";
import { AddTokensDialog } from "./add-tokens-dialog";
import type React from "react";
import { useContext, useState } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import { InputField } from "@raystack/apsara";
import type { Organization } from "@raystack/proton/frontier";

interface navConfig {
  label: string;
  onClick?: (e: React.MouseEvent<HTMLDivElement>) => void;
  disabled?: boolean;
  isLoading?: boolean;
  subItems?: navConfig[];
}

const DropdownItem = ({ item }: { item: navConfig }) => {
  return (
    <DropdownMenu.Item
      disabled={item.disabled}
      onClick={item?.onClick}
      className={styles["navbar-action-menu-item"]}
      data-test-id={`navbar-action-menu-item-${item.label}`}
    >
      <Text>{item.label}</Text>
      {item.isLoading ? <Spinner size={2} /> : null}
    </DropdownMenu.Item>
  );
};

interface NavbarActionMenuProps {
  organizationTitle: string;
  organizationId: string;
  openKYCPanel: () => void;
  openEditOrgPanel: () => void;
  openEditBillingPanel: () => void;
  onExportMembers?: () => Promise<void>;
  onExportProjects?: () => Promise<void>;
  onExportTokens?: () => Promise<void>;
}

const NavbarActionMenu = ({
  organizationId,
  openKYCPanel,
  openEditOrgPanel,
  openEditBillingPanel,
  organizationTitle,
  onExportMembers,
  onExportProjects,
  onExportTokens,
}: NavbarActionMenuProps) => {
  const [isInviteUsersDialogOpen, setIsInviteUsersDialogOpen] = useState(false);
  const [isAddTokensDialogOpen, setIsAddTokensDialogOpen] = useState(false);
  const [isMembersDownloading, setIsMembersDownloading] = useState(false);
  const [isProjectsDownloading, setIsProjectsDownloading] = useState(false);
  const [isTokensDownloading, setIsTokensDownloading] = useState(false);

  const openInviteUsersDialog = () => {
    setIsInviteUsersDialogOpen(true);
  };

  const openAddTokensDialog = () => {
    setIsAddTokensDialogOpen(true);
  };

  async function handleExportMembers(e: React.MouseEvent<HTMLDivElement>) {
    e.preventDefault();
    if (!onExportMembers) return;
    try {
      setIsMembersDownloading(true);
      await onExportMembers();
    } catch (error) {
      toast.error("Failed to export members");
      console.error(error);
    } finally {
      setIsMembersDownloading(false);
    }
  }

  async function handleExportProjects(e: React.MouseEvent<HTMLDivElement>) {
    e.preventDefault();
    if (!onExportProjects) return;
    try {
      setIsProjectsDownloading(true);
      await onExportProjects();
    } catch (error) {
      toast.error("Failed to export projects");
      console.error(error);
    } finally {
      setIsProjectsDownloading(false);
    }
  }

  async function handleExportTokens(e: React.MouseEvent<HTMLDivElement>) {
    e.preventDefault();
    if (!onExportTokens) return;
    try {
      setIsTokensDownloading(true);
      await onExportTokens();
    } catch (error) {
      toast.error("Failed to export tokens");
      console.error(error);
    } finally {
      setIsTokensDownloading(false);
    }
  }

  const exportSubItems: navConfig[] = [];
  if (onExportMembers) {
    exportSubItems.push({
      label: "Members",
      onClick: handleExportMembers,
      isLoading: isMembersDownloading,
    });
  }
  if (onExportProjects) {
    exportSubItems.push({
      label: "Projects",
      onClick: handleExportProjects,
      isLoading: isProjectsDownloading,
    });
  }
  if (onExportTokens) {
    exportSubItems.push({
      label: "Tokens",
      onClick: handleExportTokens,
      isLoading: isTokensDownloading,
    });
  }

  const items: navConfig[] = [
    {
      label: "Edit...",
      subItems: [
        { label: "Organization...", onClick: openEditOrgPanel },
        { label: "Billing...", onClick: openEditBillingPanel },
        { label: "KYC...", onClick: openKYCPanel },
      ],
    },
    {
      label: "Invite users...",
      onClick: openInviteUsersDialog,
    },
    {
      label: "Add project...",
      disabled: true,
    },
    {
      label: "Add tokens...",
      onClick: openAddTokensDialog,
    },
    {
      label: "Change plan...",
      disabled: true,
    },
    ...(exportSubItems.length > 0
      ? [{ label: "Export", subItems: exportSubItems }]
      : []),
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
          <IconButton size={2} data-test-id="admin-nav-action-menu-button">
            <DotsHorizontalIcon />
          </IconButton>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content className={styles["navbar-action-menu-content"]}>
          {items.map((item) =>
            item?.subItems ? (
              <DropdownMenu key={item.label}>
                <DropdownMenu.TriggerItem>
                  {item.label}
                </DropdownMenu.TriggerItem>
                <DropdownMenu.Content>
                  <DropdownMenu>
                    {item.subItems.map((subItem) => (
                      <DropdownItem
                        item={subItem}
                        key={`${item.label}---${subItem.label}`}
                      />
                    ))}
                  </DropdownMenu>
                </DropdownMenu.Content>
              </DropdownMenu>
            ) : (
              <DropdownItem item={item} key={item.label} />
            ),
          )}
        </DropdownMenu.Content>
      </DropdownMenu>
    </>
  );
};

const NavLinks = ({ organizationId }: { organizationId: string }) => {
  const location = useLocation();
  const currentPath = location.pathname;

  const links = [
    { name: "Members", path: `/organizations/${organizationId}/members` },
    { name: "Projects", path: `/organizations/${organizationId}/projects` },
    { name: "Invoices", path: `/organizations/${organizationId}/invoices` },
    { name: "Tokens", path: `/organizations/${organizationId}/tokens` },
    { name: "API", path: `/organizations/${organizationId}/apis` },
    // { name: "Audit log", path: `/organizations/${organizationId}/#` },
    { name: "Security", path: `/organizations/${organizationId}/security` },
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
            <Chip
              data-state={isActive ? "active" : ""}
              variant={"filled"}
              className={styles["nav-chip"]}
            >
              {link.name}
            </Chip>
          </NavLink>
        );
      })}
    </Flex>
  );
};

interface OrganizationDetailsNavbarProps {
  organization: Organization;
  toggleSidePanel: () => void;
  openKYCPanel: () => void;
  openEditOrgPanel: () => void;
  openEditBillingPanel: () => void;
  onExportMembers?: () => Promise<void>;
  onExportProjects?: () => Promise<void>;
  onExportTokens?: () => Promise<void>;
}

export const OrganizationsDetailsNavabar = ({
  organization,
  toggleSidePanel,
  openKYCPanel,
  openEditBillingPanel,
  openEditOrgPanel,
  onExportMembers,
  onExportProjects,
  onExportTokens,
}: OrganizationDetailsNavbarProps) => {
  const { search } = useContext(OrganizationContext);

  function handleSearchChange(e: React.ChangeEvent<HTMLInputElement>) {
    const value = e.target.value;
    search.onChange(value);
  }

  return (
    <nav className={styles.navbar}>
      <Flex gap={4} align="center">
        <Breadcrumb size="small">
          <Breadcrumb.Item
            href="/organizations"
            leadingIcon={<OrganizationIcon />}
          >
            Organizations
          </Breadcrumb.Item>
          <Breadcrumb.Separator />
          <Breadcrumb.Item
            href={`/organizations/${organization?.id}`}
            leadingIcon={
              <Avatar
                color={getAvatarColor(organization?.id || "")}
                src={organization?.avatar}
                fallback={organization?.title?.[0]}
                size={1}
              />
            }
          >
            {organization?.title || "NA"}
          </Breadcrumb.Item>
        </Breadcrumb>
        <NavbarActionMenu
          organizationId={organization?.id || ""}
          openKYCPanel={openKYCPanel}
          openEditOrgPanel={openEditOrgPanel}
          openEditBillingPanel={openEditBillingPanel}
          organizationTitle={organization?.title || ""}
          onExportMembers={onExportMembers}
          onExportProjects={onExportProjects}
          onExportTokens={onExportTokens}
        />
        <NavLinks organizationId={organization?.id || ""} />
      </Flex>
      <Flex align="center" gap={4}>
        {search.isVisible ? (
          <InputField
            value={search.query}
            onChange={handleSearchChange}
            placeholder="Search"
            data-test-id="admin-org-details-navbar-search"
          />
        ) : null}
        <IconButton
          size={3}
          data-test-id="admin-nav-sidepanel-button"
          onClick={toggleSidePanel}
        >
          <SidebarIcon />
        </IconButton>
      </Flex>
    </nav>
  );
};
