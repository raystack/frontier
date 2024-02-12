import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import { V1Beta1Role } from "@raystack/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { RolesHeader } from "./header";

type ContextType = { role: V1Beta1Role | null };
export default function RoleList() {
  const { client } = useFrontier();
  const [roles, setRoles] = useState([]);

  useEffect(() => {
    async function getRoles() {
      const {
        // @ts-ignore
        data: { roles },
      } = await client?.frontierServiceListRoles();
      setRoles(roles);
    }
    getRoles();
  }, []);
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
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
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
