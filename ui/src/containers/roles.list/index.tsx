import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import useSWR from "swr";
import { Role } from "~/types/role";
import { fetcher, reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { RolesHeader } from "./header";

type ContextType = { role: Role | null };
export default function RoleList() {
  const { data } = useSWR("/v1beta1/roles", fetcher);
  const { roles = [] } = data || { roles: [] };
  let { roleId } = useParams();

  const roleMapByName = reduceByKey(roles ?? [], "id");

  const tableStyle = roles?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };
  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={roles ?? []}
        // @ts-ignore
        columns={getColumns(roles)}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <RolesHeader />
          <DataTable.FilterChips style={{ paddingTop: "16px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet
            context={{
              role: roleId ? roleMapByName[roleId] : null,
            }}
          />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
  );
}

export function useRole() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 role created</h3>
    <div className="pera">Try creating a new role.</div>
  </EmptyState>
);
