import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import useSWR from "swr";
import { User } from "~/types/user";
import { fetcher, reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { UsersHeader } from "./header";

type ContextType = { user: User | null };
export default function UserList() {
  const { data } = useSWR("/v1beta1/admin/users", fetcher);
  const { users = [] } = data || { users: [] };
  let { userId } = useParams();

  const userMapByName = reduceByKey(users ?? [], "id");

  const tableStyle = users?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={users ?? []}
        // @ts-ignore
        columns={getColumns(users)}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <UsersHeader />
          <DataTable.FilterChips style={{ paddingTop: "16px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet
            context={{
              user: userId ? userMapByName[userId] : null,
            }}
          />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
  );
}

export function useUser() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 user created</h3>
    <div className="pera">Try creating a new user.</div>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
