import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import { OrganizationIcon } from "@raystack/apsara/icons";
import { useState } from "react";
import { OrganizationsNavabar } from "./navbar";
import styles from "./list.module.css";
import { getColumns } from "./columns";
import { useInfiniteQuery, useQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  FrontierServiceQueries,
  SearchOrganizationsResponse_OrganizationResult,
  type Plan,
  ListPlansRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";

import { useNavigate } from "react-router-dom";
import PageTitle from "~/components/page-title";
import { CreateOrganizationPanel } from "./create";
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage,
  DEFAULT_PAGE_SIZE,
} from "~/utils/connect-pagination";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import { useDebouncedState } from "@raystack/apsara/hooks";

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
  const [showCreatePanel, setShowCreatePanel] = useState(false);

  const [tableQuery, setTableQuery] = useDebouncedState<DataTableQuery>(
    INITIAL_QUERY,
    200,
  );

  // Transform the DataTableQuery to RQLRequest format
  const query = transformDataTableQueryToRQLRequest(tableQuery, {
    fieldNameMapping: {
      createdBy: "created_by",
      planName: "plan_name",
      subscriptionCycleEndAt: "subscription_cycle_end_at",
      paymentMode: "payment_mode",
      subscriptionState: "subscription_state",
      createdAt: "created_at",
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
    AdminServiceQueries.searchOrganizations,
    { query: query },
    {
      pageParamKey: "query",
      getNextPageParam: lastPage =>
        getConnectNextPageParam(lastPage, { query: query }, "organizations"),
      staleTime: 0,
      refetchOnWindowFocus: false,
      retry: 1,
      retryDelay: 1000,
    },
  );

  const { data: plans = [], isLoading: isPlansLoading, error: plansError } = useQuery(
    FrontierServiceQueries.listPlans,
    create(ListPlansRequestSchema, {}),
    {
      select: (data) => data?.plans || [],
    },
  );

  // Log error if it occurs
  if (plansError) {
    console.error("Failed to fetch plans:", plansError);
  }

  const data =
    infiniteData?.pages?.flatMap(page => page?.organizations || []) || [];

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
      console.error("Error loading more organizations:", error);
    }
  };

  const navigate = useNavigate();

  function closeCreateOrgPanel() {
    setShowCreatePanel(false);
  }

  function openCreateOrgPanel() {
    setShowCreatePanel(true);
  }

  const columns = getColumns({ plans, groupCountMap });

  const loading = isLoading || isPlansLoading || isFetchingNextPage;

  if (isError) {
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

  function onRowClick(row: SearchOrganizationsResponse_OrganizationResult) {
    navigate(`/organizations/${row.id}`);
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
        onRowClick={onRowClick}>
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
