import {
  DataTable,
  EmptyState,
  Flex,
  DataTableQuery,
} from "@raystack/apsara/v1";
import { V1Beta1Organization } from "@raystack/frontier";

import { useCallback, useEffect, useState } from "react";
import { OrganizationsNavabar } from "./navbar";
import OrganizationsIcon from "~/assets/icons/organization.svg?react";
import styles from "./list.module.css";
import { getColumns } from "./columns";

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

export const OrganizationList = () => {
  const [data, setData] = useState<V1Beta1Organization[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [query, setQuery] = useState<DataTableQuery>({});
  const [offset, setOffset] = useState(0);

  const columns = getColumns();

  useEffect(() => {
    async function fetchOrganizations() {
      setIsLoading(true);
    }
    fetchOrganizations();
  }, []);

  const onTableQueryChange = useCallback((newQuery: DataTableQuery) => {
    console.log(newQuery);
    setQuery(newQuery);
  }, []);

  const tableClassName =
    data.length || isLoading ? styles["table"] : styles["table-empty"];
  return (
    <DataTable
      columns={columns}
      data={data}
      isLoading={isLoading}
      defaultSort={{ name: "created_at", order: "desc" }}
      onTableQueryChange={onTableQueryChange}
      mode="server"
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
