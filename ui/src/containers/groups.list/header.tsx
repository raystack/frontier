import { PlusIcon } from "@radix-ui/react-icons";

import { DataTable, useTable } from "@raystack/apsara";
import { Button, Flex } from "@raystack/apsara/v1";
import { useNavigate } from "react-router-dom";
import PageHeader from "~/components/page-header";

const pageHeader = {
  title: "Groups",
  breadcrumb: [],
};
export const GroupsHeader = () => {
  const navigate = useNavigate();
  const { filteredColumns } = useTable();
  const isFiltered = filteredColumns.length > 0;

  return (
    <PageHeader title={pageHeader.title} breadcrumb={pageHeader.breadcrumb}>
      {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
      <DataTable.ViewOptions />
      <DataTable.GloabalSearch placeholder="Search groups..." />
      <Button
        variant="outline"
        color="neutral"
        onClick={() => navigate("/groups/create")}
        style={{ width: "100%" }}
        data-test-id="admin-ui-new-group-btn"
      >
        <Flex
          direction="column"
          align="center"
          style={{ paddingRight: "var(--pd-4)" }}
        >
          <PlusIcon />
        </Flex>
        new group
      </Button>
    </PageHeader>
  );
};
