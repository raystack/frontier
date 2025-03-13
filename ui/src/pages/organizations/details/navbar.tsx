import OrganizationsIcon from "~/assets/icons/organization.svg?react";
import {
  Flex,
  Text,
  Breadcrumb,
  Avatar,
  IconButton,
  DropdownMenu,
} from "@raystack/apsara/v1";

import styles from "./details.module.css";
import { ChevronRightIcon, DotsHorizontalIcon } from "@radix-ui/react-icons";

const NavbarActionMenu = () => {
  return (
    <DropdownMenu>
      <DropdownMenu.Trigger asChild>
        <IconButton size={2}>
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

export const OrganizationsDetailsNavabar = () => {
  return (
    <nav className={styles.navbar}>
      <Flex gap={4}>
        <Breadcrumb
          size="small"
          separator={<ChevronRightIcon style={{ display: "flex" }} />}
          items={[
            {
              label: "Organizations",
              href: "/organizations",
              icon: <OrganizationsIcon />,
            },
            {
              label: "ABC",
              href: "/organizations/abc",
              icon: <Avatar fallback="A" size={1} />,
            },
          ]}
        />
        <NavbarActionMenu />
      </Flex>
      <Flex align="center" gap={4}>
        {/* <Separator orientation="vertical" size="small" /> */}
      </Flex>
    </nav>
  );
};
