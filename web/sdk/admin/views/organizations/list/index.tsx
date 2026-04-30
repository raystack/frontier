import { DataTable, EmptyState, Flex, type DataTableQuery, type DataTableSort } from "@raystack/apsara";
import { OrganizationIcon } from "@raystack/apsara/icons";
import { useEffect, useState } from "react";
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

import { PageTitle } from "../../../components/PageTitle";
import { CreateOrganizationPanel } from "./create";
import {
  getConnectNextPageParam,
  getGroupCountMapFromFirstPage,
  DEFAULT_PAGE_SIZE,
} from "~/utils/connect-pagination";
import { transformDataTableQueryToRQLRequest } from "~/utils/transform-query";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { useDebouncedState } from "@raystack/apsara/hooks";
import { useTerminology } from "../../../hooks/useTerminology";

const NoOrganizations = () => {
  const t = useTerminology();
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading={`No ${t.organization({ case: "capital" })} Found`}
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<OrganizationIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: 'createdAt', order: 'desc' };
const INITIAL_QUERY: DataTableQuery = {
  offset: 0,
  limit: DEFAULT_PAGE_SIZE,
};

export type OrganizationListViewProps = {
  /** App name displayed in the page title (e.g. "Frontier Admin"). */
  appName?: string;
  /** Called when a user clicks on an organization row. Use to navigate to the org detail page. */
  onNavigateToOrg?: (id: string) => void;
  /** Callback to export organizations list as CSV. Shown in navbar when provided. */
  onExportCsv?: () => Promise<void>;
  /** List of allowed organization types for filtering / creation. */
  organizationTypes?: string[];
  /** Base URL of the consumer app, used for generating links. */
  appUrl?: string;
  /** List of country names for KYC country selector during org creation. */
  countries?: string[];
};

export const OrganizationListView = ({
  appName,
  onNavigateToOrg,
  onExportCsv,
  organizationTypes = [],
  appUrl,
  countries = [],
}: OrganizationListViewProps = {}) => {
  const t = useTerminology();
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
  useEffect(() => {
    if (plansError) {
      console.error("Failed to fetch plans:", plansError);
    }
  }, [plansError]);

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
        <PageTitle title={t.organization({ plural: true, case: "capital" })} appName={appName} />
        <EmptyState
          icon={<ExclamationTriangleIcon />}
          heading={`Error Loading ${t.organization({ plural: true, case: "capital" })}`}
          subHeading={
            error?.message ||
            `Something went wrong while loading ${t.organization({ plural: true, case: "lower" })}. Please try again.`
          }
        />
      </>
    );
  }

  const tableClassName =
    data.length || loading ? styles["table"] : styles["table-empty"];

  function onRowClick(row: SearchOrganizationsResponse_OrganizationResult) {
    if (row.id && onNavigateToOrg) onNavigateToOrg(row.id);
  }
  return (
    <>
      {showCreatePanel ? (
        <CreateOrganizationPanel
          onClose={closeCreateOrgPanel}
          organizationTypes={organizationTypes}
          appUrl={appUrl}
          countries={countries}
          onSuccess={(id) => {
            closeCreateOrgPanel();
            onNavigateToOrg?.(id);
          }}
        />
      ) : null}
      <PageTitle title={t.organization({ plural: true, case: "capital" })} appName={appName} />
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
            onExportCsv={onExportCsv}
          />
          <DataTable.Toolbar />
          <DataTable.VirtualizedContent
            classNames={{
              root: styles["table-wrapper"],
              table: tableClassName,
              header: styles["table-header"],
            }}
            emptyState={<NoOrganizations />}
            rowHeight={48}
            groupHeaderHeight={48}
          />
        </Flex>
      </DataTable>
    </>
  );
};
