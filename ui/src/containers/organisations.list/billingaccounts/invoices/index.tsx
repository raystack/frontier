import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { V1Beta1Invoice, V1Beta1Organization } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { OrganizationsHeader } from "../../header";
import { getColumns } from "./columns";

export default function OrganisationBAInvoices() {
  const { client } = useFrontier();
  let { organisationId, billingaccountId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [invoices, setInvoices] = useState<V1Beta1Invoice[]>([]);

  const pageHeader = {
    title: "Organizations",
    breadcrumb: [
      {
        href: `/organisations`,
        name: `Organizations list`,
      },
      {
        href: `/organisations/${organisationId}`,
        name: `${organisation?.title}`,
      },
      {
        href: `/organisations/${organisationId}/billingaccounts/${billingaccountId}`,
        name: `Billing Account`,
      },
      {
        href: "",
        name: `Invoices`,
      },
    ],
  };

  useEffect(() => {
    async function getOrganization() {
      try {
        const res = await client?.frontierServiceGetOrganization(organisationId ?? "")
        const organization = res?.data?.organization
        setOrganisation(organization);
      } catch (error) {
        console.error(error)
      }
    }
    getOrganization();
  }, [client, organisationId]);

  useEffect(() => {
    async function getOrganizationInvoices() {
      try {
        const res = await client?.frontierServiceListInvoices(organisationId ?? "", billingaccountId ?? "", { nonzero_amount_only: true })
        const invoices = res?.data?.invoices ?? []
        setInvoices(invoices);
      } catch (error) {
        console.error(error)
      }
    }
    getOrganizationInvoices();
  }, [billingaccountId, client, organisationId]);

  const tableStyle = invoices?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={invoices ?? []}
        // @ts-ignore
        columns={getColumns(invoices)}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <OrganizationsHeader header={pageHeader} />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 invoice created</h3>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
