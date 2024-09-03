import { PlusIcon } from "@radix-ui/react-icons";

import { Button, DataTable, Flex, useTable } from "@raystack/apsara";
import { useNavigate } from "react-router-dom";
import PageHeader from "~/components/page-header";

const pageHeader = {
  title: "Projects",
  breadcrumb: [],
};
export const ProjectsHeader = ({ header = pageHeader }: any) => {
  const navigate = useNavigate();
  const { filteredColumns } = useTable();
  const isFiltered = filteredColumns.length > 0;

  return (
    <PageHeader title={header.title} breadcrumb={header.breadcrumb}>
      {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
      <DataTable.ViewOptions />
      <DataTable.GloabalSearch placeholder="Search projects..." />
      <Button
        variant="secondary"
        onClick={() => navigate("/projects/create")}
        style={{ width: "100%" }}
        data-test-id="admin-ui-create-project-btn"
      >
        <Flex
          direction="column"
          align="center"
          style={{ paddingRight: "var(--pd-4)" }}
        >
          <PlusIcon />
        </Flex>
        new project
      </Button>
    </PageHeader>
  );
};
