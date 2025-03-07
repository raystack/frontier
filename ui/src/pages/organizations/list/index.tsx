import {
  DataTable,
  EmptyState,
  Flex,
  DataTableQuery,
} from "@raystack/apsara/v1";
import { V1Beta1Organization } from "@raystack/frontier";
import { useDebounceCallback } from "usehooks-ts";
import { useCallback, useEffect, useState } from "react";
import { OrganizationsNavabar } from "./navbar";
import OrganizationsIcon from "~/assets/icons/organization.svg?react";
import styles from "./list.module.css";
import { getColumns } from "./columns";
import { api } from "~/api";

const NoOrganizations = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Organization Found"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      icon={<OrganizationsIcon />}
    />
  );
};

const LIMIT = 20;
const DEFAULT_SORT = { name: "created_at", order: "desc" };

export const OrganizationList = () => {
  const [data, setData] = useState<V1Beta1Organization[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [query, setQuery] = useState<DataTableQuery>({});
  const [nextOffset, setNextOffset] = useState(0);

  const columns = getColumns();

  const fetchOrganizations = useCallback(async (apiQuery = {}) => {
    try {
      setIsLoading(true);
      const response = await api.adminServiceSearchOrganizations({
        limit: LIMIT,
        ...apiQuery,
      });
      const organizations = response.data.organizations || [];
      setData((prev) => [...prev, ...organizations]);
      setNextOffset(response.data.pagination?.offset || 0);
    } catch (error) {
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchOrganizations({ offset: 0, sort: [DEFAULT_SORT] });
  }, [fetchOrganizations]);

  async function fetchMoreOrganizations() {
    fetchOrganizations({ offset: nextOffset + LIMIT, ...query });
  }

  const onTableQueryChange = useDebounceCallback((newQuery: DataTableQuery) => {
    setData([]);
    fetchOrganizations({ ...newQuery, offset: 0 });
    setQuery(newQuery);
  }, 500);

  const tableClassName =
    data.length || isLoading ? styles["table"] : styles["table-empty"];
  return (
    <DataTable
      columns={columns}
      data={data}
      isLoading={isLoading}
      defaultSort={DEFAULT_SORT as any}
      onTableQueryChange={onTableQueryChange}
      mode="server"
      onLoadMore={fetchMoreOrganizations}
    >
      <Flex direction="column" style={{ width: "100%" }}>
        <OrganizationsNavabar seachQuery={query.search} />
        <DataTable.Toolbar />
        <DataTable.Content
          classNames={{
            table: tableClassName,
          }}
          emptyState={<NoOrganizations />}
        />
      </Flex>
    </DataTable>
  );
};
