import { DataTable, EmptyState } from "@raystack/apsara";
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
  return (
    <DataTable
      data={roles ?? []}
      // @ts-ignore
      columns={getColumns(roles)}
      emptyState={noDataChildren}
      style={{
        width: "100%",
        maxHeight: "calc(100vh - 60px)",
        overflow: "scroll",
      }}
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
