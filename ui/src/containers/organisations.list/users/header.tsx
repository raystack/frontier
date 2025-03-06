import { PlusIcon } from "@radix-ui/react-icons";

import { Button, Flex } from "@raystack/apsara/v1";
import { DataTable, useTable } from "@raystack/apsara";
import { useNavigate } from "react-router-dom";
import PageHeader from "~/components/page-header";

const defaultPageHeader = {
  title: "Organizations",
  breadcrumb: [],
};

export const OrganizationsUsersHeader = ({
  header = defaultPageHeader,
  orgId,
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
        <DataTable.GloabalSearch placeholder="Search users..." />
        <Button
          variant="outline"
          color="neutral"
          style={{ width: "100%" }}
          onClick={() => navigate(`/organisations/${orgId}/users/invite`)}
          data-test-id="admin-ui-invite-users-btn"
        >
          <Flex
            direction="column"
            align="center"
            style={{ paddingRight: "var(--pd-4)" }}
          >
            <PlusIcon />
          </Flex>
          Invite users
        </Button>
      </PageHeader>
    </>
  );
};
