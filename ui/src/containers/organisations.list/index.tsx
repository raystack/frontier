import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useContext } from "react";
import { Outlet, useOutletContext } from "react-router-dom";

import { V1Beta1Organization } from "@raystack/frontier";
import { getColumns } from "./columns";
import { OrganizationsHeader } from "./header";
import { AppContext } from "~/contexts/App";

type ContextType = { organisation: V1Beta1Organization | null };
export default function OrganisationList() {
  const { organizations, isLoading } = useContext(AppContext);

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
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 organisation created</h3>
    <div className="pera">Try creating a new organisation.</div>
  </EmptyState>
);
