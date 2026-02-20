import {
  DataTable,
  Flex,
  Text,
  Separator,
  IconButton,
  Spinner,
} from "@raystack/apsara";
import styles from "./list.module.css";
import { DownloadIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import UserIcon from "../../../assets/icons/UsersIcon";
import React, { useState } from "react";
import { InviteUser } from "./invite-users";

interface NavbarProps {
  searchQuery?: string;
  onExportUsers?: () => Promise<void>;
}

const Navbar = ({ searchQuery, onExportUsers }: NavbarProps) => {
  const [showSearch, setShowSearch] = useState(searchQuery ? true : false);
  const [isDownloading, setIsDownloading] = useState(false);

  function toggleSearch() {
    setShowSearch((prev) => !prev);
  }

  function onSearchBlur(e: React.FocusEvent<HTMLInputElement>) {
    const value = e.target.value;
    if (!value) {
      setShowSearch(false);
    }
  }

  async function onDownloadClick() {
    if (!onExportUsers) return;
    try {
      setIsDownloading(true);
      await onExportUsers();
    } catch (error) {
      console.error(error);
    } finally {
      setIsDownloading(false);
    }
  }

  return (
    <nav className={styles.navbar}>
      <Flex gap={2}>
        <UserIcon />
        <Text size={2} weight={500}>
          Users
        </Text>
      </Flex>
      <Flex align="center" gap={4}>
        <InviteUser />
        <Separator orientation="vertical" size="small" />
        {showSearch ? (
          <DataTable.Search
            autoFocus
            showClearButton={true}
            size="small"
            onBlur={onSearchBlur}
          />
        ) : (
          <IconButton
            size={3}
            aria-label="Search"
            data-test-id="admin-search-users-btn"
            onClick={toggleSearch}
          >
            <MagnifyingGlassIcon />
          </IconButton>
        )}
        <IconButton
          size={3}
          aria-label="Download"
          data-test-id="admin-download-users-list-btn"
          onClick={onDownloadClick}
          disabled={isDownloading || !onExportUsers}
        >
          {isDownloading ? <Spinner /> : <DownloadIcon />}
        </IconButton>
      </Flex>
    </nav>
  );
};

export default Navbar;
