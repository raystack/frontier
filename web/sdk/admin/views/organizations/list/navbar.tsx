import {
  Button,
  DataTable,
  Flex,
  Text,
  Separator,
  IconButton,
  Spinner,
} from "@raystack/apsara";
import { OrganizationIcon } from "@raystack/apsara/icons";
import styles from "./list.module.css";
import {
  DownloadIcon,
  MagnifyingGlassIcon,
  PlusIcon,
} from "@radix-ui/react-icons";
import React, { useState } from "react";

interface OrganizationsNavabarProps {
  searchQuery?: string;
  openCreatePanel: () => void;
  onExportCsv?: () => Promise<void>;
}

export const OrganizationsNavabar = ({
  searchQuery,
  openCreatePanel,
  onExportCsv,
}: OrganizationsNavabarProps) => {
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
    if (!onExportCsv) return;
    try {
      setIsDownloading(true);
      await onExportCsv();
    } catch (error) {
      console.error(error);
    } finally {
      setIsDownloading(false);
    }
  }

  return (
    <nav className={styles.navbar}>
      <Flex gap={2}>
        <OrganizationIcon />
        <Text size={2} weight={500}>
          Organizations
        </Text>
      </Flex>
      <Flex align="center" gap={4}>
        <Button
          variant="text"
          color="neutral"
          leadingIcon={<PlusIcon />}
          data-test-id="admin-create-organization-btn"
          onClick={openCreatePanel}
        >
          New Organization
        </Button>
        <Separator orientation="vertical" size="small" />
        {showSearch ? (
          <DataTable.Search
            showClearButton={true}
            size="small"
            onBlur={onSearchBlur}
          />
        ) : (
          <IconButton
            size={3}
            aria-label="Search"
            data-test-id="admin-search-organization-btn"
            onClick={toggleSearch}
          >
            <MagnifyingGlassIcon />
          </IconButton>
        )}
        {onExportCsv ? (
          <IconButton
            size={3}
            aria-label="Download"
            data-test-id="admin-download-organization-list-btn"
            onClick={onDownloadClick}
            disabled={isDownloading}
          >
            {isDownloading ? <Spinner /> : <DownloadIcon />}
          </IconButton>
        ) : null}
      </Flex>
    </nav>
  );
};
