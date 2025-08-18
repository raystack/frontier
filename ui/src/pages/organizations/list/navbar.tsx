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
import { exportCsvFromStream } from "~/utils/helper";
import { clients } from "~/connect/clients";

const adminClient = clients.admin({ useBinary: true });

interface OrganizationsNavabarProps {
  searchQuery?: string;
  openCreatePanel: () => void;
}

export const OrganizationsNavabar = ({
  searchQuery,
  openCreatePanel,
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
    try {
      setIsDownloading(true);
      await exportCsvFromStream(
        adminClient.exportOrganizations,
        {},
        "organizations.csv"
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
          data-test-id="admin-ui-create-organization-btn"
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
            data-test-id="admin-ui-search-organization-btn"
            onClick={toggleSearch}
          >
            <MagnifyingGlassIcon />
          </IconButton>
        )}
        <IconButton
          size={3}
          aria-label="Download"
          data-test-id="admin-ui-download-organization-list-btn"
          onClick={onDownloadClick}
          disabled={isDownloading}
        >
          {isDownloading ? <Spinner /> : <DownloadIcon />}
        </IconButton>
      </Flex>
    </nav>
  );
};
