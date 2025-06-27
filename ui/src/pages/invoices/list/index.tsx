import {
  DataTable,
  type DataTableQuery,
  type DataTableSort,
  EmptyState,
  Flex,
} from "@raystack/apsara/v1";
import PageTitle from "~/components/page-title";
import { InvoicesNavabar } from "./navbar";
import styles from "./list.module.css";
import InvoicesIcon from "~/assets/icons/invoices.svg?react";
import { getColumns } from "./columns";
import { api } from "~/api";
import { useCallback } from "react";
import { useRQL } from "~/hooks/useRQL";
import type { V1Beta1SearchInvoicesResponseInvoice } from "~/api/frontier";

const NoInvoices = () => {
  return (
    <EmptyState
      classNames={{
        container: styles["empty-state"],
        subHeading: styles["empty-state-subheading"],
      }}
      heading="No invoices found"
      subHeading="Start billing to organizations to populate the table"
      icon={<InvoicesIcon />}
    />
  );
};

const DEFAULT_SORT: DataTableSort = { name: "created_at", order: "desc" };

export const InvoicesList = () => {
  const columns = getColumns();

  const apiCallback = useCallback(
    async (apiQuery: DataTableQuery = {}) =>
      await api.adminServiceSearchInvoices(apiQuery).then((res) => res.data),
    [],
  );

  const {
    data,
    loading: isLoading,
    query,
    onTableQueryChange,
    fetchMore,
  } = useRQL<V1Beta1SearchInvoicesResponseInvoice>({
    initialQuery: { offset: 0, sort: [DEFAULT_SORT] },
    dataKey: "invoices",
    key: "invoices",
    fn: apiCallback,
    onError: (error: Error | unknown) =>
      console.error("Failed to fetch invoices:", error),
  });

  return (
    <>
      <PageTitle title="Invoices" />
      <DataTable
        columns={columns}
        data={data}
        isLoading={isLoading}
        defaultSort={DEFAULT_SORT}
        onTableQueryChange={onTableQueryChange}
        mode="server"
        onLoadMore={fetchMore}
      >
        <Flex direction="column" style={{ width: "100%" }}>
          <InvoicesNavabar />
          <DataTable.Toolbar />
          <DataTable.Content
            classNames={{
              root: styles["table-wrapper"],
              header: styles["table-header"],
            }}
            emptyState={<NoInvoices />}
          />
        </Flex>
      </DataTable>
    </>
  );
};
