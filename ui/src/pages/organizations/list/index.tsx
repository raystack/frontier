import {
  DataTable,
  EmptyState,
  Flex,
  DataTableQuery,
  EmptyFilterValue,
} from "@raystack/apsara/v1";
import { OrganizationIcon } from "@raystack/apsara/icons";
import { V1Beta1Organization, V1Beta1Plan } from "@raystack/frontier";
import { useDebounceCallback } from "usehooks-ts";
import { useCallback, useEffect, useState } from "react";
import { OrganizationsNavabar } from "./navbar";
import styles from "./list.module.css";
import { getColumns } from "./columns";
import { api } from "~/api";
import { useNavigate } from "react-router-dom";
import PageTitle from "~/components/page-title";

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

const LIMIT = 50;
const DEFAULT_SORT = { name: "created_at", order: "desc" };

export const OrganizationList = () => {
  const [data, setData] = useState<V1Beta1Organization[]>([]);
  const [isDataLoading, setIsDataLoading] = useState(false);

  const [plans, setPlans] = useState<V1Beta1Plan[]>([]);
  const [isPlansLoading, setIsPlansLoading] = useState(false);

  const [query, setQuery] = useState<DataTableQuery>({});
  const [nextOffset, setNextOffset] = useState(0);
  const [hasMoreData, setHasMoreData] = useState(true);
  const [groupCountMap, setGroupCountMap] = useState<
    Record<string, Record<string, number>>
  >({});

  const naviagte = useNavigate();

  const fetchOrganizations = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      try {
        setIsDataLoading(true);
        const response = await api.adminServiceSearchOrganizations({
          ...apiQuery,
          filters:
            apiQuery.filters?.map((fil) => ({
              ...fil,
              operator: fil.value === EmptyFilterValue ? "empty" : fil.operator,
            })) || [],
          limit: LIMIT,
        });
        const organizations = response.data.organizations || [];
        setData((prev) => [...prev, ...organizations]);
        setNextOffset(response.data.pagination?.offset || 0);
        const groupCount =
          response.data.group?.data?.reduce(
            (acc, group) => {
              acc[group.name || ""] = group.count || 0;
              return acc;
            },
            {} as Record<string, number>,
          ) || {};
        const groupKey = response.data.group?.name;
        if (groupKey) {
          setGroupCountMap((prev) => ({ ...prev, [groupKey]: groupCount }));
        }
        setHasMoreData(
          organizations.length !== 0 && organizations.length === LIMIT,
        );
      } catch (error) {
        console.error(error);
      } finally {
        setIsDataLoading(false);
      }
    },
    [],
  );

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
    fetchOrganizations({ offset: 0, sort: [DEFAULT_SORT as any] });
    fetchPlans();
  }, [fetchOrganizations, fetchPlans]);

  async function fetchMoreOrganizations() {
    if (isDataLoading || !hasMoreData) {
      return;
    }
    fetchOrganizations({ offset: nextOffset + LIMIT, ...query });
  }

  const onTableQueryChange = useDebounceCallback((newQuery: DataTableQuery) => {
    setData([]);
    fetchOrganizations({ ...newQuery, offset: 0 });
    setQuery(newQuery);
  }, 500);

  const columns = getColumns({ plans, groupCountMap: groupCountMap });

  const isLoading = isDataLoading || isPlansLoading;

  const tableClassName =
    data.length || isLoading ? styles["table"] : styles["table-empty"];

  function onRowClick(row: V1Beta1Organization) {
    naviagte(`/organisations/${row.id}`);
  }
  return (
    <>
      <PageTitle title="Organizations" />
      <DataTable
        columns={columns}
        data={data}
        isLoading={isLoading}
        defaultSort={DEFAULT_SORT as any}
        onTableQueryChange={onTableQueryChange}
        mode="server"
        onLoadMore={fetchMoreOrganizations}
        onRowClick={onRowClick}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <OrganizationsNavabar seachQuery={query.search} />
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
