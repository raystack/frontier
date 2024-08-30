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

const DEFAULT_PAGE_SIZE = 200;

type ContextType = { user: V1Beta1User | null };
export default function UserList() {
  const { client } = useFrontier();
  const [users, setUsers] = useState<V1Beta1User[]>([]);
  const [isUsersLoading, setIsUsersLoading] = useState(false);

  useEffect(() => {
    async function getAllUsers() {
      setIsUsersLoading(true);
      try {
        const {
          // @ts-ignore
          data: { users },
        } = await client?.adminServiceListAllUsers({
          page_size: DEFAULT_PAGE_SIZE,
        }) || {};
        setUsers(users);
      } catch (err) {
        console.error(err);
      } finally {
        setIsUsersLoading(false);
      }
    }
    getAllUsers();
  }, [client]);

  let { userId } = useParams();

  const userMapByName = reduceByKey(users ?? [], "id");

  const tableStyle = users?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const userList = isUsersLoading
    ? [...new Array(5)].map((_, i) => ({
        name: i.toString(),
        title: "",
      }))
    : users;

  const columns = getColumns();

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={userList ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
        isLoading={isUsersLoading}
      >
        <DataTable.Toolbar>
          <UsersHeader header={pageHeader} />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
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
    <h3>No users created</h3>
    <div className="pera">Try creating a new user.</div>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
