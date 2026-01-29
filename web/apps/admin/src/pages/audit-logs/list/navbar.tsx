import { DataTable, Flex, Text, IconButton, Spinner } from "@raystack/apsara";
import CpuChipIcon from "~/assets/icons/cpu-chip.svg?react";
import styles from "./list.module.css";
import { DownloadIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import React, { useCallback, useState } from "react";
import { clients } from "~/connect/clients";
import { exportCsvFromStream } from "~/utils/helper";
import { useQueryClient } from "@tanstack/react-query";
import { AUDIT_LOG_QUERY_KEY } from "../util";
import { RQLExportRequest } from "@raystack/proton/frontier";

const adminClient = clients.admin({ useBinary: true });

interface NavbarProps {
  searchQuery?: string;
}

const Navbar = ({ searchQuery }: NavbarProps) => {
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
    try {
      setIsDownloading(true);
      const query = queryClient.getQueryData(
        AUDIT_LOG_QUERY_KEY,
      ) as RQLExportRequest;

      await exportCsvFromStream(
        adminClient.exportAuditRecords,
        { query },
        "audit-logs.csv",
      );
    } catch (error) {
      console.error(error);
    } finally {
      setIsDownloading(false);
    }
  }, [queryClient]);

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
        <IconButton
          size={3}
          aria-label="Download"
          data-test-id="admin-download-audit-logs-list-btn"
          onClick={onDownloadClick}
          disabled={isDownloading}>
          {isDownloading ? <Spinner /> : <DownloadIcon />}
        </IconButton>
      </Flex>
    </nav>
  );
};

export default Navbar;
