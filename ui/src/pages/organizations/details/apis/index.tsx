import { DataTable, EmptyState, Flex } from "@raystack/apsara/v1";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara/v1";
import styles from "./apis.module.css";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { useCallback, useContext, useEffect } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import PageTitle from "~/components/page-title";
import { getColumns } from "./columns";
import { api } from "~/api";
import type { SearchOrganizationServiceUsersResponseOrganizationServiceUser } from "~/api/frontier";
import { useRQL } from "~/hooks/useRQL";

const NoCredentials = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No API Credentials found"
      subHeading="We couldn't find any matches for that keyword or filter. Try alternative terms or check for typos."
      // TODO: update icon with raystack icon
      icon={<InfoCircledIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };

export function OrganizationApisPage() {
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";

  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const title = `API | ${organization?.title} | Organizations`;

  const apiCallback = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      const response = await api?.adminServiceSearchOrganizationServiceUsers(
        organizationId,
        { ...apiQuery, search: searchQuery || "" },
      );
      return response?.data;
    },
    [organizationId, searchQuery],
  );

  const { data, loading, query, onTableQueryChange, groupCountMap, fetchMore } =
    useRQL<SearchOrganizationServiceUsersResponseOrganizationServiceUser>({
      initialQuery: { offset: 0 },
      key: organizationId,
      dataKey: "organization_service_users",
      fn: apiCallback,
      searchParam: searchQuery || "",
      onError: (error: Error | unknown) =>
        console.error("Failed to fetch service users:", error),
    });

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  const isLoading = loading;

  const columns = getColumns({ groupCountMap });
  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
      <DataTable
        columns={columns}
        data={data}
        isLoading={isLoading}
        defaultSort={DEFAULT_SORT}
        mode="server"
        onTableQueryChange={onTableQueryChange}
        onLoadMore={fetchMore}
        query={{ ...query, search: searchQuery }}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={<NoCredentials />}
            classNames={{
              table: styles["table"],
              root: styles["table-wrapper"],
              header: styles["table-header"],
            }}
          />
        </Flex>
      </DataTable>
    </Flex>
  );
}
