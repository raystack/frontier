import { DataTable, Flex, Text, IconButton, Spinner } from "@raystack/apsara";
import CpuChipIcon from "~/assets/icons/cpu-chip.svg?react";
import styles from "./list.module.css";
import { DownloadIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import React, { useState } from "react";
import { clients } from "~/connect/clients";
import { exportCsvFromStream } from "~/utils/helper";

const adminClient = clients.admin({ useBinary: true });

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
      await exportCsvFromStream(
        adminClient.exportAuditRecords,
        {},
        "audit-logs.csv",
      );
    } catch (error) {
      console.error(error);
    } finally {
      setIsDownloading(false);
    }
  }

  return (
    <nav className={styles.navbar}>
      <Flex gap={2}>
        <CpuChipIcon />
        <Text size={2} weight={500}>
          Audit Logs
        </Text>
      </Flex>
      <Flex align="center" gap={4}>
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
            data-test-id="admin-ui-search-audit-logs-btn"
            onClick={toggleSearch}>
            <MagnifyingGlassIcon />
          </IconButton>
        )}
        <IconButton
          size={3}
          aria-label="Download"
          data-test-id="admin-ui-download-audit-logs-list-btn"
          onClick={onDownloadClick}
          disabled={isDownloading}>
          {isDownloading ? <Spinner /> : <DownloadIcon />}
        </IconButton>
      </Flex>
    </nav>
  );
};

export default Navbar;
