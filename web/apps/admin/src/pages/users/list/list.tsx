import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import { useState } from "react";
import Navbar from "./navbar";
import styles from "./list.module.css";
import { getColumns } from "./columns";
import { useNavigate } from "react-router-dom";
import PageTitle from "~/components/page-title";
import UserIcon from "~/assets/icons/users.svg?react";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries, type User } from "@raystack/proton/frontier";
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage,
  DEFAULT_PAGE_SIZE,
  transformDataTableQueryToRQLRequest,
} from "@raystack/frontier/admin";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useDebouncedState } from "@raystack/apsara/hooks";

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
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};

export const UsersList = () => {
  const navigate = useNavigate();
  const [tableQuery, setTableQuery] = useDebouncedState<DataTableQuery>(
    INITIAL_QUERY,
    200,
  );

  // Transform the DataTableQuery to RQLRequest format
  const query = transformDataTableQueryToRQLRequest(tableQuery, {
    fieldNameMapping: {
      createdAt: "created_at",
      updatedAt: "updated_at",
    },
  });

  const {
    data: infiniteData,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    error,
    isError,
  } = useInfiniteQuery(
    AdminServiceQueries.searchUsers,
    { query: query },
    {
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(lastPage, { query: query }, "users"),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const data = infiniteData?.pages?.flatMap(page => page?.users || []) || [];

  const groupCountMap = infiniteData
    ? getGroupCountMapFromFirstPage(infiniteData)
    : {};

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery({
      ...newQuery,
      offset: 0,
      limit: newQuery.limit || DEFAULT_PAGE_SIZE,
    });
  };

  const handleLoadMore = async () => {
    try {
      await fetchNextPage();
    } catch (error) {
      console.error("Error loading more users:", error);
    }
  };

  const columns = getColumns({ groupCountMap });

  const loading = isLoading || isFetchingNextPage;

  const onRowClick = (row: User) => {
    navigate(`/users/${row.id}`);
  };

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <>
        <PageTitle title="Users" />
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="Error Loading Users"
          subHeading={
            error?.message ||
            "Something went wrong while loading users. Please try again."
          }
        />
      </>
    );
  }

  const tableClassName =
    data.length || loading ? styles["table"] : styles["table-empty"];

  return (
    <>
      <PageTitle title="Users" />
      <DataTable
        query={tableQuery}
        columns={columns}
        data={data}
        isLoading={loading}
        defaultSort={DEFAULT_SORT}
        onTableQueryChange={onTableQueryChange}
        mode="server"
        onLoadMore={handleLoadMore}
        onRowClick={onRowClick}>
        <Flex direction="column" style={{ width: "100%" }}>
          <Navbar searchQuery={tableQuery.search} />
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
