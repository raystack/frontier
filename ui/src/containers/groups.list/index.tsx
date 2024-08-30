import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { V1Beta1Group } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useContext, useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { GroupsHeader } from "./header";
import { AppContext } from "~/contexts/App";

type ContextType = { group: V1Beta1Group | null };
export default function GroupList() {
  const { client } = useFrontier();
  const { orgMap } = useContext(AppContext);
  const [groups, setGroups] = useState<V1Beta1Group[]>([]);
  const [isGroupsLoading, setIsGroupsLoading] = useState(false);

  useEffect(() => {
    async function getGroups() {
      setIsGroupsLoading(true);
      try {
        const {
          // @ts-ignore
          data: { groups },
        } = await client?.adminServiceListGroups() ?? {};
        setGroups(groups);
      } catch (err) {
        console.error(err);
      } finally {
        setIsGroupsLoading(false);
      }
    }
    getGroups();
  }, [client]);

  const groupMapByName = reduceByKey(groups ?? [], "id");
  let { groupId } = useParams();

  const tableStyle = groups?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const groupList = isGroupsLoading
    ? [...new Array(5)].map((_, i) => ({
        name: i.toString(),
        title: "",
      }))
    : groups;

  const columns = getColumns({
    orgMap,
  });

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={groupList ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
        isLoading={isGroupsLoading}
      >
        <DataTable.Toolbar>
          <GroupsHeader />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
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
