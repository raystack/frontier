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

import styles from "./details.module.css";
import {
  ChevronRightIcon,
  DotsHorizontalIcon,
  MagnifyingGlassIcon,
} from "@radix-ui/react-icons";
import { V1Beta1Organization } from "~/api/frontier";

const NavbarActionMenu = () => {
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
        <DropdownMenu.Item>
          <Text>Invite users...</Text>
        </DropdownMenu.Item>
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
              href: "/organizations",
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
        <Flex gap={3}>
          <Chip>Members</Chip>
          <Chip>Projects</Chip>
          <Chip>Tokens</Chip>
          <Chip>API</Chip>
          <Chip>Audit log</Chip>
          <Chip>Security</Chip>
        </Flex>
      </Flex>
      <Flex align="center" gap={4}>
        <IconButton size={3} data-test-id="admin-ui-nav-search-button">
          <MagnifyingGlassIcon />
        </IconButton>
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
