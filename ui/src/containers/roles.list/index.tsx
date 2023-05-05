import { EmptyState, Flex, Table } from "@odpf/apsara";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import useSWR from "swr";
import { tableStyle } from "~/styles";
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
  console.log(roleMapByName);
  return (
    <Flex direction="row" css={{ height: "100%", width: "100%" }}>
      <Table
        css={tableStyle}
        columns={getColumns(roles)}
        data={roles ?? []}
        noDataChildren={noDataChildren}
      >
        <Table.TopContainer>
          <RolesHeader />
        </Table.TopContainer>
        <Table.DetailContainer
          css={{
            borderLeft: "1px solid $gray4",
            borderTop: "1px solid $gray4",
          }}
        >
          <Outlet
            context={{
              role: roleId ? roleMapByName[roleId] : null,
            }}
          />
        </Table.DetailContainer>
      </Table>
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
