import { PlusIcon } from "@radix-ui/react-icons";
import { DataTable, useTable } from "@raystack/apsara";
import { Button, Flex } from "@raystack/apsara/v1";
import { useNavigate } from "react-router-dom";
import PageHeader from "~/components/page-header";

export const UsersHeader = ({ header }: any) => {
  const navigate = useNavigate();
  const { filteredColumns } = useTable();
  const isFiltered = filteredColumns.length > 0;

  return (
    <PageHeader title={header.title} breadcrumb={header?.breadcrumb || []}>
      {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
      <DataTable.ViewOptions />
      <DataTable.GloabalSearch placeholder="Search users..." />
      <Button
        color="neutral"
        variant="outline"
        onClick={() => navigate("/users/create")}
        style={{ width: "100%" }}
        data-test-id="admin-ui-user-create-btn"
      >
        <Flex
          direction="column"
          align="center"
          style={{ paddingRight: "var(--pd-4)" }}
        >
          <PlusIcon />
        </Flex>
        new user
      </Button>
    </PageHeader>
  );
};
