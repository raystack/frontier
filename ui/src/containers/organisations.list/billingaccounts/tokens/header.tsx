import { PlusIcon } from "@radix-ui/react-icons";

import { DataTable, useTable } from "@raystack/apsara";
import { Button, Flex } from "@raystack/apsara/v1";

import { useNavigate, useParams } from "react-router-dom";
import PageHeader from "~/components/page-header";

const defaultPageHeader = {
  title: "Organizations",
  breadcrumb: [],
};

export const OrganizationsTokenHeader = ({
  header = defaultPageHeader,
  ...props
}: any) => {
  const navigate = useNavigate();
  const { filteredColumns } = useTable();
  const isFiltered = filteredColumns.length > 0;
  let { organisationId, billingaccountId } = useParams();

  return (
    <>
      <PageHeader
        title={header.title}
        breadcrumb={header.breadcrumb}
        {...props}
      >
        {isFiltered ? <DataTable.ClearFilter /> : <DataTable.FilterOptions />}
        <DataTable.ViewOptions />
        <DataTable.GloabalSearch placeholder="Search transaction..." />
        <Button
          variant={"outline"}
          color="neutral"
          onClick={() =>
            navigate(
              `/organisations/${organisationId}/billingaccounts/${billingaccountId}/tokens/add`
            )
          }
          style={{ width: "100%" }}
          data-test-id="admin-ui-add-tokens-btn"
        >
          <Flex
            direction="column"
            align="center"
            style={{ paddingRight: "var(--pd-4)" }}
          >
            <PlusIcon />
          </Flex>
          Add tokens
        </Button>
      </PageHeader>
    </>
  );
};
