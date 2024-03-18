import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { V1Beta1Invoice } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useContext, useEffect, useState } from "react";
import PageHeader from "~/components/page-header";
import { getColumns } from "./columns";
import { AppContext } from "~/contexts/App";

const pageHeader = {
  title: "Invoices",
  breadcrumb: [],
};

export default function InvoicesList() {
  const { client } = useFrontier();
  const { orgMap } = useContext(AppContext);
  const [invoices, setInvoices] = useState<V1Beta1Invoice[]>([]);
  const [billingOrgMap, setBillingOrgMap] = useState<Record<string, string>>(
    {}
  );
  const [isInvoicesLoading, setIsInvoicesLoading] = useState(false);

  const columns = getColumns({
    billingOrgMap,
    orgMap,
    isLoading: isInvoicesLoading,
  });

  const tableStyle = invoices?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  useEffect(() => {
    async function getInvoices() {
      try {
        setIsInvoicesLoading(true);
        const [invoicesResp, billingAccountResp] = await Promise.all([
          client?.adminServiceListAllInvoices(),
          client?.adminServiceListAllBillingAccounts(),
        ]);
        const invoiceList = invoicesResp?.data?.invoices || [];
        setInvoices(invoiceList);

        const billingAccounts =
          billingAccountResp?.data?.billing_accounts || [];
        const billingIdOrgMap = billingAccounts.reduce((acc, ba) => {
          const id = ba.id || "";
          acc[id] = ba?.org_id || "";
          return acc;
        }, {} as Record<string, string>);
        setBillingOrgMap(billingIdOrgMap);
      } finally {
        setIsInvoicesLoading(false);
      }
    }
    getInvoices();
  }, [client]);

  const invoicesList = isInvoicesLoading
    ? [...new Array(5)].map((_, i) => ({
        id: i.toString(),
      }))
    : invoices;

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={invoicesList}
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
