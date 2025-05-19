import { DataTable, EmptyState, Flex } from "@raystack/apsara/v1";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara/v1";
import { OrganizationIcon } from "@raystack/apsara/icons";
import { useCallback, useEffect, useState } from "react";
import { OrganizationsNavabar } from "./navbar";
import styles from "./list.module.css";
import { getColumns } from "./columns";
import { api } from "~/api";
import { useNavigate } from "react-router-dom";
import PageTitle from "~/components/page-title";
import { CreateOrganizationPanel } from "./create";
import { useRQL } from "~/hooks/useRQL";
import type {
  SearchOrganizationsResponseOrganizationResult,
  V1Beta1Organization,
  V1Beta1Plan,
} from "~/api/frontier";

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

export const OrganizationList = () => {
  const [plans, setPlans] = useState<V1Beta1Plan[]>([]);
  const [isPlansLoading, setIsPlansLoading] = useState(false);

  const [showCreatePanel, setShowCreatePanel] = useState(false);

  const apiCallback = useCallback(async (apiQuery: DataTableQuery = {}) => {
    const response = await api.adminServiceSearchOrganizations({ ...apiQuery });
    return response?.data;
  }, []);

  const { data, loading, query, onTableQueryChange, fetchMore, groupCountMap } =
    useRQL<SearchOrganizationsResponseOrganizationResult>({
      key: "organizations",
      initialQuery: { offset: 0 },
      dataKey: "organizations",
      fn: apiCallback,
      onError: (error: Error | unknown) =>
        console.error("Failed to fetch orgs:", error),
    });

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

  const isLoading = loading || isPlansLoading;

  const tableClassName =
    data.length || isLoading ? styles["table"] : styles["table-empty"];

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
        columns={columns}
        data={data}
        isLoading={isLoading}
        defaultSort={DEFAULT_SORT}
        onTableQueryChange={onTableQueryChange}
        mode="server"
        onLoadMore={fetchMore}
        onRowClick={onRowClick}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <OrganizationsNavabar
            searchQuery={query.search}
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
