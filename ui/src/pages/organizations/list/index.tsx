import { DataTable, EmptyState, Flex } from "@raystack/apsara/v1";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara/v1";
import { OrganizationIcon } from "@raystack/apsara/icons";
import { useCallback, useEffect, useState } from "react";
import { OrganizationsNavabar } from "./navbar";
import styles from "./list.module.css";
import { getColumns } from "./columns";
import { api } from "~/api";
import { useInfiniteQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";

import { useNavigate } from "react-router-dom";
import PageTitle from "~/components/page-title";
import { CreateOrganizationPanel } from "./create";
import type { V1Beta1Organization, V1Beta1Plan } from "~/api/frontier";
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage,
  DEFAULT_PAGE_SIZE,
} from "~/utils/connect-pagination";

const NoOrganizations = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Organization Found"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<OrganizationIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};

export const OrganizationList = () => {
  const [plans, setPlans] = useState<V1Beta1Plan[]>([]);
  const [isPlansLoading, setIsPlansLoading] = useState(false);

  const [showCreatePanel, setShowCreatePanel] = useState(false);

  const [tableQuery, setTableQuery] = useState<DataTableQuery>(INITIAL_QUERY);

  const {
    data: infiniteData,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    error,
    isError,
  } = useInfiniteQuery(
    AdminServiceQueries.searchOrganizations,
    { query: tableQuery },
    {
      pageParamKey: "query",
      getNextPageParam: (lastPage) =>
        getConnectNextPageParam(
          lastPage,
          { query: tableQuery },
          "organizations",
        ),
      staleTime: 0,
      refetchOnWindowFocus: false,
    },
  );

  const data =
    infiniteData?.pages.flatMap((page) => page.organizations || []) || [];

  const groupCountMap = getGroupCountMapFromFirstPage(infiniteData);

  const onTableQueryChange = (newQuery: DataTableQuery) => {
    setTableQuery({
      ...newQuery,
      offset: 0,
      limit: newQuery.limit || DEFAULT_PAGE_SIZE,
    });
  };

  const handleLoadMore = async () => {
    await fetchNextPage();
  };

  const naviagte = useNavigate();

  function closeCreateOrgPanel() {
    setShowCreatePanel(false);
  }

  function openCreateOrgPanel() {
    setShowCreatePanel(true);
  }

  const fetchPlans = useCallback(async () => {
    try {
      setIsPlansLoading(true);
      const response = await api.frontierServiceListPlans();
      const newPlans = response.data.plans || [];
      setPlans(newPlans);
    } catch (error) {
      console.error(error);
    } finally {
      setIsPlansLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchPlans();
  }, [fetchPlans]);

  const columns = getColumns({ plans, groupCountMap: groupCountMap });

  const loading = isLoading || isPlansLoading || isFetchingNextPage;

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <>
        <PageTitle title="Organizations" />
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading="Error Loading Organizations"
          subHeading={
            error?.message ||
            "Something went wrong while loading organizations. Please try again."
          }
        />
      </>
    );
  }

  const tableClassName =
    data.length || loading ? styles["table"] : styles["table-empty"];

  function onRowClick(row: V1Beta1Organization) {
    naviagte(`/organizations/${row.id}`);
  }
  return (
    <>
      {showCreatePanel ? (
        <CreateOrganizationPanel onClose={closeCreateOrgPanel} />
      ) : null}
      <PageTitle title="Organizations" />
      <DataTable
        query={tableQuery}
        columns={columns}
        data={data}
        isLoading={loading}
        defaultSort={DEFAULT_SORT}
        onTableQueryChange={onTableQueryChange}
        mode="server"
        onLoadMore={handleLoadMore}
        onRowClick={onRowClick}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <OrganizationsNavabar
            searchQuery={tableQuery.search}
            openCreatePanel={openCreateOrgPanel}
          />
          <DataTable.Toolbar />
          <DataTable.Content
            classNames={{
              root: styles["table-wrapper"],
              table: tableClassName,
              header: styles["table-header"],
            }}
            emptyState={<NoOrganizations />}
          />
        </Flex>
      </DataTable>
    </>
  );
};
