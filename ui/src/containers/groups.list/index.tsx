import { EmptyState, Flex, Table } from "@odpf/apsara";
import { Outlet } from "react-router-dom";
import useSWR from "swr";
import { fetcher, tableStyle } from "~/utils/helper";
import { columns } from "./columns";

export default function GroupList() {
  const { data, error } = useSWR("/admin/v1beta1/groups", fetcher);
  const { groups = [] } = data || { groups: [] };
  return (
    <Flex direction="row" css={{ height: "100%", width: "100%" }}>
      <Table
        css={tableStyle}
        columns={columns}
        data={groups ?? []}
        noDataChildren={noDataChildren}
      ></Table>
      <Outlet />
    </Flex>
  );
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 group created</h3>
    <div className="pera">Try creating a new group.</div>
  </EmptyState>
);
