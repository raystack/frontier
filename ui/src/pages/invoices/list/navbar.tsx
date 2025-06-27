import { DataTable, Flex, IconButton, Text } from "@raystack/apsara/v1";
import styles from "./list.module.css";
import InvoicesIcon from "~/assets/icons/invoices.svg?react";
import { useState } from "react";
import { MagnifyingGlassIcon } from "@radix-ui/react-icons";

export const InvoicesNavabar = ({ searchQuery }: { searchQuery: string }) => {
  const [showSearch, setShowSearch] = useState(searchQuery ? true : false);
  function toggleSearch() {
    setShowSearch((prev) => !prev);
  }

  function onSearchBlur(e: React.FocusEvent<HTMLInputElement>) {
    const value = e.target.value;
    if (!value) {
      setShowSearch(false);
    }
  }

  return (
    <nav className={styles.navbar}>
      <Flex gap={2}>
        <InvoicesIcon />
        <Text size={2} weight={500}>
          Invoices
        </Text>
      </Flex>
      <Flex align="center">
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
            data-test-id="admin-ui-search-invoices-btn"
            onClick={toggleSearch}
          >
            <MagnifyingGlassIcon />
          </IconButton>
        )}
      </Flex>
    </nav>
  );
};
