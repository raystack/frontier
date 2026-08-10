import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableSort } from "@raystack/apsara";
import Navbar from "./navbar";
import styles from "./list.module.css";
import { getColumns } from "./columns";
import { PageTitle } from "../../../components/PageTitle";
import UserIcon from "../../../assets/icons/UsersIcon";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries, type User } from "@raystack/proton/frontier";
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage,
} from "~/utils/connect-pagination";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useTerminology } from "../../../hooks/useTerminology";
import { useServerTableQuery } from "../../../hooks/useServerTableQuery";

const NoUsers = () => {
  const t = useTerminology();
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading={`No ${t.user({ plural: true, case: "capital" })} Found`}
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<UserIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: 'createdAt', order: 'desc' };
const TRANSFORM_OPTIONS = {
  fieldNameMapping: {
    createdAt: "created_at",
    updatedAt: "updated_at",
  },
};

interface UsersListProps {
  onExportUsers?: () => Promise<void>;
  onNavigateToUser?: (userId: string) => void;
}

export const UsersList = ({ onExportUsers, onNavigateToUser }: UsersListProps) => {
  const t = useTerminology();
  const {
    tableQuery,
    rqlQuery: query,
    onTableQueryChange,
  } = useServerTableQuery({
    defaultSort: DEFAULT_SORT,
    transformOptions: TRANSFORM_OPTIONS,
  });

  const {
    data: infiniteData,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
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

  const handleLoadMore = async () => {
    if (!hasNextPage || isFetchingNextPage) return;
    try {
      await fetchNextPage();
    } catch (error) {
      console.error("Error loading more users:", error);
    }
  };

  const columns = getColumns({ groupCountMap, onNavigateToUser });

  const loading = isLoading || isFetchingNextPage;

  const onRowClick = (row: User) => {
    onNavigateToUser?.(row.id);
  };

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <>
        <PageTitle title={t.user({ plural: true, case: "capital" })} />
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading={`Error Loading ${t.user({ plural: true, case: "capital" })}`}
          subHeading={
            error?.message ||
            `Something went wrong while loading ${t.user({ plural: true, case: "lower" })}. Please try again.`
          }
        />
      </>
    );
  }

  const tableClassName =
    data.length || loading ? styles["table"] : styles["table-empty"];

  return (
    <>
      <PageTitle title={t.user({ plural: true, case: "capital" })} />
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
          <Navbar searchQuery={tableQuery.search} onExportUsers={onExportUsers} />
          <DataTable.Toolbar />
          <DataTable.VirtualizedContent
            classNames={{
              root: styles["table-wrapper"],
              table: tableClassName,
              header: styles["table-header"],
            }}
            emptyState={<NoUsers />}
            rowHeight={48}
            groupHeaderHeight={48}
          />
        </Flex>
      </DataTable>
    </>
  );
};
