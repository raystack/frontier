import { EmptyState, Flex, Table } from "@raystack/apsara";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import useSWR from "swr";
import { tableStyle } from "~/styles";
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

  return (
    <Flex direction="row" css={{ height: "100%", width: "100%" }}>
      <Table
        css={tableStyle}
        columns={getColumns(groups)}
        data={groups ?? []}
        noDataChildren={noDataChildren}
      >
        <Table.TopContainer>
          <GroupsHeader />
        </Table.TopContainer>
        <Table.DetailContainer
          css={{
            borderLeft: "1px solid $gray4",
            borderTop: "1px solid $gray4",
          }}
        >
          <Outlet
            context={{ group: groupId ? groupMapByName[groupId] : null }}
          />
        </Table.DetailContainer>
      </Table>
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
