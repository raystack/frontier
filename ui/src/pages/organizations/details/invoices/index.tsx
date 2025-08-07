import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import type { DataTableQuery, DataTableSort } from "@raystack/apsara";
import styles from "./invoices.module.css";
import { FileTextIcon } from "@radix-ui/react-icons";
import { useCallback, useContext, useEffect } from "react";
import { OrganizationContext } from "../contexts/organization-context";
import PageTitle from "~/components/page-title";
import { getColumns } from "./columns";
import { api } from "~/api";
import type { SearchOrganizationInvoicesResponseOrganizationInvoice } from "~/api/frontier";
import { useRQL } from "~/hooks/useRQL";

const NoInvoices = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No Invoice found"
      subHeading="We couldn’t find any matches for that keyword or filter. Try alternative terms or check for typos."
      // TODO: update icon with raystack icon
      icon={<FileTextIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };

export function OrganizationInvoicesPage() {
  const { organization, search } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";

  const {
    onChange: onSearchChange,
    setVisibility: setSearchVisibility,
    query: searchQuery,
  } = search;

  const title = `Invoices | ${organization?.title} | Organizations`;

  const apiCallback = useCallback(
    async (apiQuery: DataTableQuery = {}) => {
      const response = await api?.adminServiceSearchOrganizationInvoices(
        organizationId,
        apiQuery,
      );
      return response?.data;
    },
    [organizationId],
  );

  useEffect(() => {
    setSearchVisibility(true);
    return () => {
      onSearchChange("");
      setSearchVisibility(false);
    };
  }, [setSearchVisibility, onSearchChange]);

  const { data, loading, query, onTableQueryChange, fetchMore, groupCountMap } =
    useRQL<SearchOrganizationInvoicesResponseOrganizationInvoice>({
      initialQuery: { offset: 0 },
      key: organizationId,
      dataKey: "organization_invoices",
      fn: apiCallback,
      searchParam: searchQuery || "",
      onError: (error: Error | unknown) =>
        console.error("Failed to fetch invoices:", error),
    });

  const columns = getColumns({ groupCountMap });
  return (
    <Flex justify="center" className={styles["container"]}>
      <PageTitle title={title} />
      <DataTable
        columns={columns}
        data={data}
        isLoading={loading}
        defaultSort={DEFAULT_SORT}
        mode="server"
        onTableQueryChange={onTableQueryChange}
        onLoadMore={fetchMore}
        query={{ ...query, search: searchQuery }}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <DataTable.Toolbar />
          <DataTable.Content
            emptyState={<NoInvoices />}
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
