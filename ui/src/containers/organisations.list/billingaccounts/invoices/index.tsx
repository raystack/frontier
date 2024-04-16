import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { V1Beta1Organization } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { OrganizationsHeader } from "../../header";
import { getColumns } from "./columns";

export default function OrganisationBAInvoices() {
  const { client } = useFrontier();
  let { organisationId, billingaccountId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [invoices, setInvoices] = useState([]);

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
      const {
        // @ts-ignore
        data: { organization },
      } = await client?.frontierServiceGetOrganization(organisationId ?? "");
      setOrganisation(organization);
    }
    getOrganization();
  }, [client, organisationId]);

  useEffect(() => {
    async function getOrganizationInvoices() {
      const {
        // @ts-ignore
        data: { invoices },
      } = await client?.frontierServiceListInvoices(
        organisationId ?? "",
        billingaccountId ?? "",
        { nonzero_amount_only: true }
      );
      setInvoices(invoices);
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
