import { EmptyState, Flex, Table } from "@odpf/apsara";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import useSWR from "swr";
import { tableStyle } from "~/styles";
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
  console.log(userMapByName);
  return (
    <Flex direction="row" css={{ height: "100%", width: "100%" }}>
      <Table
        css={tableStyle}
        columns={getColumns(users)}
        data={users ?? []}
        noDataChildren={noDataChildren}
      >
        <Table.TopContainer>
          <UsersHeader />
        </Table.TopContainer>
        <Table.DetailContainer
          css={{
            borderLeft: "1px solid $gray4",
            borderTop: "1px solid $gray4",
          }}
        >
          <Outlet
            context={{
              user: userId ? userMapByName[userId] : null,
            }}
          />
        </Table.DetailContainer>
      </Table>
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
