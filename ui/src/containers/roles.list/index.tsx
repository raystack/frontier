import { EmptyState, Flex } from "@raystack/apsara/v1";
import { DataTable } from "@raystack/apsara";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import { V1Beta1Role } from "@raystack/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { RolesHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";

type ContextType = { role: V1Beta1Role | null };
export default function RoleList() {
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [isRolesLoading, setIsRolesLoading] = useState(false);

  useEffect(() => {
    async function getRoles() {
      setIsRolesLoading(true);
      try {
        const res = await api?.frontierServiceListRoles();
        const roles = res?.data?.roles ?? [];
        setRoles(roles);
      } catch (err) {
        console.log(err);
      } finally {
        setIsRolesLoading(false);
      }
    }
    getRoles();
  }, []);
  let { roleId } = useParams();
  const roleMapByName = reduceByKey(roles ?? [], "id");

  const tableStyle = roles?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const roleList = isRolesLoading
    ? [...new Array(5)].map((_, i) => ({
        name: i.toString(),
        title: "",
      }))
    : roles;

  const columns = getColumns();
  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={roleList ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
        isLoading={isRolesLoading}
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
  <EmptyState icon={<ExclamationTriangleIcon />} heading="0 role created" />
);
