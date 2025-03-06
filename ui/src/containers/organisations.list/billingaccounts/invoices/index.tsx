import { DataTable } from "@raystack/apsara";
import { EmptyState, Flex } from "@raystack/apsara/v1";

import { V1Beta1Invoice, V1Beta1Organization } from "@raystack/frontier";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { OrganizationsHeader } from "../../header";
import { getColumns } from "./columns";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";

export default function OrganisationBAInvoices() {
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
        const res = await api?.frontierServiceGetOrganization(
          organisationId ?? ""
        );
        const organization = res?.data?.organization;
        setOrganisation(organization);
      } catch (error) {
        console.error(error);
      }
    }
    getOrganization();
  }, [organisationId]);

  useEffect(() => {
    async function getOrganizationInvoices() {
      try {
        const res = await api?.frontierServiceListInvoices(
          organisationId ?? "",
          billingaccountId ?? "",
          { nonzero_amount_only: true }
        );
        const invoices = res?.data?.invoices ?? [];
        setInvoices(invoices);
      } catch (error) {
        console.error(error);
      }
    }
    getOrganizationInvoices();
  }, [billingaccountId, organisationId]);

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
  <EmptyState icon={<ExclamationTriangleIcon />} heading="0 invoice created" />
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
