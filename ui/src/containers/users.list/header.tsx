import { PlusIcon } from "@radix-ui/react-icons";
import { Button, DataTable, Flex, useTable } from "@raystack/apsara";
import { useNavigate } from "react-router-dom";
import PageHeader from "~/components/page-header";

const pageHeader = {
  title: "Users",
  breadcrumb: [],
};

export const UsersHeader = () => {
  const navigate = useNavigate();
  const { filteredColumns, table } = useTable();
  const isFiltered = filteredColumns.length > 0;

  return (
    <PageHeader title={pageHeader.title} breadcrumb={pageHeader.breadcrumb}>
      {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
      <DataTable.ViewOptions />
      <DataTable.GloabalSearch placeholder="Search users..." />
      <Button
        variant="secondary"
        onClick={() => navigate("/users/create")}
        style={{ width: "100%" }}
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
