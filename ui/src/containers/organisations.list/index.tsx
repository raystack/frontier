import { DataTable } from "@raystack/apsara";
import { Flex, EmptyState } from "@raystack/apsara/v1";
import { useContext } from "react";
import { Outlet, useOutletContext } from "react-router-dom";

import { V1Beta1Organization } from "@raystack/frontier";
import { getColumns } from "./columns";
import { OrganizationsHeader } from "./header";
import { AppContext } from "~/contexts/App";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

type ContextType = { organisation: V1Beta1Organization | null };
export default function OrganisationList() {
  const { organizations, isLoading, loadMoreOrganizations } =
    useContext(AppContext);

  const tableStyle = organizations?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const columns = getColumns();
  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={organizations ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
        isLoading={isLoading}
        onLoadMore={loadMoreOrganizations}
      >
        <DataTable.Toolbar>
          <OrganizationsHeader />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
  );
}

export function useOrganisation() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState
    heading="0 organisation created"
    subHeading="Try creating a new organisation."
    icon={<ExclamationTriangleIcon />}
  />
);
