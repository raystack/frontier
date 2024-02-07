import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import { V1Beta1User } from "@raystack/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { UsersHeader } from "./header";

const pageHeader = {
  title: "Users",
  breadcrumb: [],
};

const DEFAULT_PAGE_SIZE = 200

type ContextType = { user: V1Beta1User | null };
export default function UserList() {
  const { client } = useFrontier();
  const [users, setUsers] = useState([]);

  useEffect(() => {
    async function getAllUsers() {
      const {
        // @ts-ignore
        data: { users },
      } = await client?.adminServiceListAllUsers({
        page_size: DEFAULT_PAGE_SIZE
      });
      setUsers(users);
    }
    getAllUsers();
  }, []);

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
          <UsersHeader header={pageHeader} />
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
