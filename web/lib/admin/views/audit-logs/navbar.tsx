import { DataTable, Flex, Text, IconButton, Spinner } from "@raystack/apsara";
import { CpuChipIcon } from "../../assets/icons/CpuChipIcon";
import styles from "./audit-logs.module.css";
import { DownloadIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import React, { useCallback, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AUDIT_LOG_QUERY_KEY } from "./util";
import { RQLRequest } from "@raystack/proton/frontier";

interface NavbarProps {
  searchQuery?: string;
  onExportCsv?: (query: RQLRequest) => Promise<void>;
}

const Navbar = ({ searchQuery, onExportCsv }: NavbarProps) => {
  const [showSearch, setShowSearch] = useState(searchQuery ? true : false);
  const [isDownloading, setIsDownloading] = useState(false);
  const queryClient = useQueryClient();

  const toggleSearch = useCallback(() => {
    setShowSearch(prev => !prev);
  }, []);

  const onSearchBlur = useCallback((e: React.FocusEvent<HTMLInputElement>) => {
    const value = e.target.value;
    if (!value) {
      setShowSearch(false);
    }
  }, []);

  const onDownloadClick = useCallback(async () => {
    if (!onExportCsv) return;
    try {
      setIsDownloading(true);
      const query = queryClient.getQueryData(
        AUDIT_LOG_QUERY_KEY,
      ) as RQLRequest;
      await onExportCsv(query);
    } catch (error) {
      console.error(error);
    } finally {
      setIsDownloading(false);
    }
  }, [queryClient, onExportCsv]);

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
            data-test-id="admin-search-audit-logs-btn"
            onClick={toggleSearch}>
            <MagnifyingGlassIcon />
          </IconButton>
        )}
        {onExportCsv && (
        <IconButton
          size={3}
          aria-label="Download"
          data-test-id="admin-download-audit-logs-list-btn"
          onClick={onDownloadClick}
          disabled={isDownloading}>
          {isDownloading ? <Spinner /> : <DownloadIcon />}
        </IconButton>
        )}
      </Flex>
    </nav>
  );
};

export default Navbar;
