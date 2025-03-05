import { PlusIcon } from "@radix-ui/react-icons";

import { Button, Flex } from "@raystack/apsara/v1";
import { DataTable, useTable } from "@raystack/apsara";
import { useNavigate } from "react-router-dom";
import PageHeader from "~/components/page-header";

const defaultPageHeader = {
  title: "Organizations",
  breadcrumb: [],
};

export const OrganizationsHeader = ({
  header = defaultPageHeader,
  ...props
}: any) => {
  const navigate = useNavigate();
  const { filteredColumns } = useTable();
  const isFiltered = filteredColumns.length > 0;

  return (
    <>
      <PageHeader
        title={header.title}
        breadcrumb={header.breadcrumb}
        {...props}
      >
        {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
        <DataTable.ViewOptions />
        <DataTable.GloabalSearch placeholder="Search organisations..." />
        <Button
          variant="outline"
          color="neutral"
          onClick={() => navigate("/organisations/create")}
          style={{ width: "100%" }}
          data-test-id="admin-ui-add-new-organisation-btn"
        >
          <Flex
            direction="column"
            align="center"
            style={{ paddingRight: "var(--pd-4)" }}
          >
            <PlusIcon />
          </Flex>
          new organisation
        </Button>
      </PageHeader>
    </>
  );
};
