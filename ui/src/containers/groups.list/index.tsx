import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { V1Beta1Group } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { GroupsHeader } from "./header";

type ContextType = { group: V1Beta1Group | null };
export default function GroupList() {
  const { client } = useFrontier();
  const [groups, setGroups] = useState([]);
  useEffect(() => {
    async function getGroups() {
      const {
        // @ts-ignore
        data: { groups },
      } = await client?.adminServiceListGroups();
      setGroups(groups);
    }
    getGroups();
  }, []);

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
