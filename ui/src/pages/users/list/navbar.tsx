import {
  DataTable,
  Flex,
  Text,
  Separator,
  IconButton,
  Spinner,
} from "@raystack/apsara/v1";
import UserIcon from "~/assets/icons/users.svg?react";
import styles from "./list.module.css";
import { DownloadIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import React, { useState } from "react";
import { api } from "~/api";
import { downloadFile } from "~/utils/helper";
import { InviteUser } from "./invite-users";

interface NavbarProps {
  searchQuery?: string;
}

const Navbar = ({ searchQuery }: NavbarProps) => {
  const [showSearch, setShowSearch] = useState(searchQuery ? true : false);
  const [isDownloading, setIsDownloading] = useState(false);

  function toggleSearch() {
    setShowSearch(prev => !prev);
  }

  function onSearchBlur(e: React.FocusEvent<HTMLInputElement>) {
    const value = e.target.value;
    if (!value) {
      setShowSearch(false);
    }
  }

  async function onDownloadClick() {
    try {
      setIsDownloading(true);
      const response = await api.adminServiceExportUsers({
        format: "blob",
      });
      downloadFile(response.data, "users.csv");
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
            data-test-id="admin-ui-search-users-btn"
            onClick={toggleSearch}>
            <MagnifyingGlassIcon />
          </IconButton>
        )}
        <IconButton
          size={3}
          aria-label="Download"
          data-test-id="admin-ui-download-users-list-btn"
          onClick={onDownloadClick}
          disabled={isDownloading}>
          {isDownloading ? <Spinner /> : <DownloadIcon />}
        </IconButton>
      </Flex>
    </nav>
  );
};

export default Navbar;
