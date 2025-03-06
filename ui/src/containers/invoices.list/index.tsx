import { DataTable } from "@raystack/apsara";
import { EmptyState, Flex } from "@raystack/apsara/v1";

import { V1Beta1Invoice } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useContext, useEffect, useState } from "react";
import { getColumns } from "./columns";
import { AppContext } from "~/contexts/App";
import { InvoicesHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

const pageHeader = {
  title: "Invoices",
  breadcrumb: [],
};

// TODO: Setting this to 1000 initially till APIs support filters and sorting.
const page_size = 1000;
const page_num = 1;

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
  });

  const tableStyle = invoices?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  useEffect(() => {
    async function getInvoices() {
      try {
        setIsInvoicesLoading(true);
        const [invoicesResp, billingAccountResp] = await Promise.all([
          client?.adminServiceListAllInvoices({ page_num, page_size }),
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
        isLoading={isInvoicesLoading}
      >
        <DataTable.Toolbar>
          <InvoicesHeader header={pageHeader} />

          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
}

export const noDataChildren = (
  <EmptyState icon={<ExclamationTriangleIcon />} heading="No Invoices" />
);
