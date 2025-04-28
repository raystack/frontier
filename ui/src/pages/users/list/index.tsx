import {
  DataTable,
  EmptyState,
  Flex,
  DataTableQuery,
  DataTableSort,
} from "@raystack/apsara/v1";
import { V1Beta1User } from "@raystack/frontier";
import { useCallback } from "react";
import Navbar from "./navbar";
import styles from "./list.module.css";
import { getColumns } from "./columns";
import { api } from "~/api";
import { useNavigate } from "react-router-dom";
import PageTitle from "~/components/page-title";
import UserIcon from "~/assets/icons/users.svg?react";
import { useRQL } from "~/hooks/useRQL";

const NoUsers = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Users Found"
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<UserIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };

export const UsersList = () => {
  const navigate = useNavigate();

  const apiCallback = useCallback(
    async (apiQuery: DataTableQuery = {}) =>
      await api.adminServiceSearchUsers(apiQuery).then(res => res.data),
    [],
  );

  const { data, loading, query, onTableQueryChange, fetchMore, groupCountMap } =
    useRQL<V1Beta1User>({
      initialQuery: { offset: 0, sort: [DEFAULT_SORT] },
      dataKey: "users",
      resourceId: "users",
      fn: apiCallback,
      onError: (error: Error | unknown) =>
        console.error("Failed to fetch users:", error),
    });

  const columns = getColumns({ groupCountMap });
  const isLoading = loading;

  const tableClassName =
    data.length || isLoading ? styles["table"] : styles["table-empty"];

  function onRowClick(row: V1Beta1User) {
    navigate(`/users/${row.id}`);
  }

  return (
    <>
      <PageTitle title="Users" />
      <DataTable
        columns={columns}
        data={data}
        isLoading={isLoading}
        defaultSort={DEFAULT_SORT}
        onTableQueryChange={onTableQueryChange}
        mode="server"
        onLoadMore={fetchMore}
        onRowClick={onRowClick}>
        <Flex direction="column" style={{ width: "100%" }}>
          <Navbar searchQuery={query.search} />
          <DataTable.Toolbar />
          <DataTable.Content
            classNames={{
              root: styles["table-wrapper"],
              table: tableClassName,
              header: styles["table-header"],
            }}
            emptyState={<NoUsers />}
          />
        </Flex>
      </DataTable>
    </>
  );
};
