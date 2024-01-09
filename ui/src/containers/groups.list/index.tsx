import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import useSWR from "swr";
import { Group } from "~/types/group";
import { fetcher, reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { GroupsHeader } from "./header";

type ContextType = { group: Group | null };
export default function GroupList() {
  const { data, error } = useSWR("/v1beta1/admin/groups", fetcher);
  const { groups = [] } = data || { groups: [] };
  const groupMapByName = reduceByKey(groups ?? [], "id");
  let { groupId } = useParams();

  const tableStyle = groups?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={groups ?? []}
        // @ts-ignore
        columns={getColumns(groups)}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <GroupsHeader />
          <DataTable.FilterChips style={{ paddingTop: "16px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet
            context={{
              group: groupId ? groupMapByName[groupId] : null,
            }}
          />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
  );
}

export function useGroup() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 group created</h3>
    <div className="pera">Try creating a new group.</div>
  </EmptyState>
);
