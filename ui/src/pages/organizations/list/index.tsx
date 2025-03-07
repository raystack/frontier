import {
  DataTable,
  EmptyState,
  Flex,
  Link,
  DataTableQuery,
  DataTableColumnDef,
} from "@raystack/apsara/v1";
import { V1Beta1Organization } from "@raystack/frontier";
import dayjs from "dayjs";

import { useCallback, useEffect, useState } from "react";
import { OrganizationsNavabar } from "./navbar";
import OrganizationsIcon from "~/assets/icons/organization.svg?react";
import styles from "./list.module.css";

const getColumns = (): DataTableColumnDef<V1Beta1Organization, unknown>[] => {
  return [
    {
      accessorKey: "id",
      header: "Name",
      columnType: "text",
      cell: ({ row }) => {
        return (
          <Link href={`/organisations/${row.getValue("id")}`}>
            {row.original.title}
          </Link>
        );
      },
    },
    {
      accessorKey: "created_at",
      header: "Created At",
      columnType: "date",
      cell: ({ row }) => {
        return dayjs(row.original.created_at).format("YYYY-MM-DD");
      },
    },
  ];
};

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

export const OrganizationList = () => {
  const [data, setData] = useState<V1Beta1Organization[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [query, setQuery] = useState<DataTableQuery>({});

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

  return (
    <DataTable
      columns={columns}
      data={data}
      // isLoading={isLoading}
      defaultSort={{ name: "created_at", order: "desc" }}
      onTableQueryChange={onTableQueryChange}
      mode="server"
    >
      <Flex direction="column" style={{ width: "100%" }}>
        <OrganizationsNavabar seachQuery={query.search} />
        <DataTable.Content
          classNames={{
            table: styles["table"],
          }}
          emptyState={<NoOrganizations />}
        />
      </Flex>
    </DataTable>
  );
};
