import { MagnifyingGlassIcon } from "@radix-ui/react-icons";
import { IconButton, Search } from "@raystack/apsara/v1";
import React, { useState } from "react";

interface CollapsableSearchProps {
  size?: string;
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

export const CollapsableSearch = ({
  size = "small",
  value = "",
  onChange = () => {},
}: CollapsableSearchProps) => {
  const [showSeach, setShowSearch] = useState(value ? true : false);

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
    <div>
      {showSeach ? (
        <Search
          showClearButton={true}
          size={size}
          onBlur={onSearchBlur}
          onChange={onChange}
          data-test-id="admin-ui-search-input"
        />
      ) : (
        <IconButton
          size={3}
          aria-label="Search"
          data-test-id="admin-ui-search-btn"
          onClick={toggleSearch}
        >
          <MagnifyingGlassIcon />
        </IconButton>
      )}
    </div>
  );
};
