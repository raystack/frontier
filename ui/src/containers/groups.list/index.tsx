import { DataTable } from "@raystack/apsara";
import { EmptyState, Flex } from "@raystack/apsara/v1";

import { V1Beta1Group } from "@raystack/frontier";
import { useContext, useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { GroupsHeader } from "./header";
import { AppContext } from "~/contexts/App";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";

type ContextType = { group: V1Beta1Group | null };
export default function GroupList() {
  const { orgMap } = useContext(AppContext);
  const [groups, setGroups] = useState<V1Beta1Group[]>([]);
  const [isGroupsLoading, setIsGroupsLoading] = useState(false);

  useEffect(() => {
    async function getGroups() {
      setIsGroupsLoading(true);
      try {
        const res = await api?.adminServiceListGroups();
        const groups = res?.data.groups ?? [];
        setGroups(groups);
      } catch (err) {
        console.error(err);
      } finally {
        setIsGroupsLoading(false);
      }
    }
    getGroups();
  }, []);

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
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="0 group created"
    subHeading="Try creating a new group."
  />
);
