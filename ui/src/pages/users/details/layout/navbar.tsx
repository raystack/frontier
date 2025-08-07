import {
  Avatar,
  Breadcrumb,
  Chip,
  Flex,
  IconButton,
  getAvatarColor,
} from "@raystack/apsara";
import { NavLink } from "react-router-dom";
import { ChevronRightIcon } from "@radix-ui/react-icons";
import { SidebarIcon } from "@raystack/apsara/icons";
import UserIcon from "~/assets/icons/users.svg?react";
import styles from "./navbar.module.css";
import { getUserName } from "../../util";
import { useUser } from "../user-context";

interface UserDetailsNavbarProps {
  toggleSidePanel: () => void;
}

export const UserDetailsNavbar = ({
  toggleSidePanel,
}: UserDetailsNavbarProps) => {
  const { user } = useUser();

  const links = [
    { name: "Audit log", path: `/users/${user?.id}/audit-log` },
    { name: "Security", path: `/users/${user?.id}/security` },
  ];

  return (
    <nav className={styles.navbar}>
      <Flex gap="medium" align="center">
        <Breadcrumb size="small">
          <Breadcrumb.Item href="/users" leadingIcon={<UserIcon />}>
            Users
          </Breadcrumb.Item>
          <Breadcrumb.Separator />
          <Breadcrumb.Item
            href={`/users/${user?.id}`}
            leadingIcon={
              <Avatar
                color={getAvatarColor(user?.id || "")}
                src={user?.avatar}
                fallback={getUserName(user)?.[0]}
                size={1}
              />
            }
          >
            {getUserName(user)}
          </Breadcrumb.Item>
        </Breadcrumb>
        <Flex gap="small">
          {links.map((link, index) => (
            <NavLink to={link.path} key={link.path + index}>
              {({ isActive }) => (
                <Chip
                  data-state={isActive ? "active" : undefined}
                  variant="filled"
                  className={styles["nav-chip"]}
                >
                  {link.name}
                </Chip>
              )}
            </NavLink>
          ))}
        </Flex>
      </Flex>
      <Flex align="center" gap={4}>
        <IconButton
          size={3}
          data-test-id="admin-ui-user-details-sidepanel-button"
          onClick={toggleSidePanel}
        >
          <SidebarIcon />
        </IconButton>
      </Flex>
    </nav>
  );
};
