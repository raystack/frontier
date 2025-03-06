import { DataTable } from "@raystack/apsara";
import { EmptyState, Flex } from "@raystack/apsara/v1";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import { V1Beta1User } from "@raystack/frontier";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { UsersHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

const pageHeader = {
  title: "Users",
  breadcrumb: [],
};

const DEFAULT_PAGE_SIZE = 2000;

type ContextType = { user: V1Beta1User | null };
export default function UserList() {
  const { client } = useFrontier();
  const [users, setUsers] = useState<V1Beta1User[]>([]);
  const [isUsersLoading, setIsUsersLoading] = useState(false);

  useEffect(() => {
    async function getAllUsers() {
      setIsUsersLoading(true);
      try {
        const res = await client?.adminServiceListAllUsers({
          page_size: DEFAULT_PAGE_SIZE,
        });
        const users = res?.data?.users ?? [];
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
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="No users created"
    subHeading="Try creating a new user."
  />
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
