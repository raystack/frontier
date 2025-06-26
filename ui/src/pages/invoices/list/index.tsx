import { DataTable, EmptyState, Flex } from "@raystack/apsara/v1";
import PageTitle from "~/components/page-title";
import { InvoicesNavabar } from "./navbar";
import styles from "./list.module.css";
import InvoicesIcon from "~/assets/icons/invoices.svg?react";

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

export const InvoicesList = () => {
  return (
    <>
      <PageTitle title="Invoices" />
      <DataTable
        columns={[]}
        data={[]}
        isLoading={false}
        defaultSort={{}}
        // onTableQueryChange={onTableQueryChange}
        // mode="server"
        // onLoadMore={fetchMore}
        // onRowClick={onRowClick}
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
