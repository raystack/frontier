import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect } from "react";
import PageHeader from "~/components/page-header";

const pageHeader = {
  title: "Invoices",
  breadcrumb: [],
};

export default function InvoicesList() {
  // @ts-ignore
  const columns = [];
  const invoices = [];
  const data = [];

  const { client } = useFrontier();

  const tableStyle = invoices?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  useEffect(() => {
    async function getInvoices() {
      const resp = await client?.adminServiceListAllInvoices();
      console.log(resp);
    }
    getInvoices();
  }, [client]);

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        // @ts-ignore
        data={data}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <PageHeader
            title={pageHeader.title}
            breadcrumb={pageHeader.breadcrumb}
          />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <div className="pera">No Invoices</div>
  </EmptyState>
);
