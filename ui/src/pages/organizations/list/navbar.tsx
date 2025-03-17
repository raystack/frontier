import OrganizationsIcon from "~/assets/icons/organization.svg?react";
import {
  Button,
  DataTable,
  Flex,
  Text,
  Separator,
  IconButton,
  Spinner,
} from "@raystack/apsara/v1";

import styles from "./list.module.css";
import {
  DownloadIcon,
  MagnifyingGlassIcon,
  PlusIcon,
} from "@radix-ui/react-icons";
import React, { useState } from "react";
import { api } from "~/api";

interface OrganizationsNavabarProps {
  seachQuery?: string;
}

const downloadFile = (data: File, filename: string) => {
  const link = document.createElement("a");
  const downloadUrl = window.URL.createObjectURL(new Blob([data]));
  link.href = downloadUrl;
  link.setAttribute("download", filename);
  document.body.appendChild(link);
  link.click();
  link.parentNode?.removeChild(link);
  window.URL.revokeObjectURL(downloadUrl);
};

export const OrganizationsNavabar = ({
  seachQuery,
}: OrganizationsNavabarProps) => {
  const [showSeach, setShowSearch] = useState(seachQuery ? true : false);
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
      const response = await api.adminServiceExportOrganizations({
        format: "blob",
      });
      downloadFile(response.data, "organizations.csv");
    } catch (error) {
      console.error(error);
    } finally {
      setIsDownloading(false);
    }
  }

  return (
    <nav className={styles.navbar}>
      <Flex gap={2}>
        <OrganizationsIcon />
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
        >
          New Organization
        </Button>
        <Separator orientation="vertical" size="small" />
        {showSeach ? (
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
