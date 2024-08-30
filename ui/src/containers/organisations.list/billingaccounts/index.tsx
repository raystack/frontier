import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { V1Beta1BillingAccount, V1Beta1Organization } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { reduceByKey } from "~/utils/helper";
import { OrganizationsHeader } from "../header";
import { getColumns } from "./columns";

type ContextType = { billingaccount: V1Beta1BillingAccount | null };
export default function OrganisationBillingAccounts() {
  const { client } = useFrontier();
  let { organisationId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [billingAccounts, setBillingAccounts] = useState([]);

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
        href: ``,
        name: `Billing Accounts`,
      },
    ],
  };

  useEffect(() => {
    async function getOrganization() {
      const {
        // @ts-ignore
        data: { organization },
      } = await client?.frontierServiceGetOrganization(organisationId ?? "") ?? {};
      setOrganisation(organization);
    }
    getOrganization();
  }, [client, organisationId]);

  useEffect(() => {
    async function getOrganizationBillingAccounts() {
      const {
        // @ts-ignore
        data: { billing_accounts },
      } = await client?.frontierServiceListBillingAccounts(
        organisationId ?? ""
      ) ?? {};
      setBillingAccounts(billing_accounts);
    }
    getOrganizationBillingAccounts();
  }, [client, organisationId]);

  let { billingaccountId } = useParams();
  const billingAccountsMapByName = reduceByKey(billingAccounts ?? [], "id");

  const tableStyle = billingAccounts?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={billingAccounts ?? []}
        // @ts-ignore
        columns={getColumns(billingAccounts)}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <OrganizationsHeader header={pageHeader} />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet
            context={{
              billingaccount: billingaccountId
                ? billingAccountsMapByName[billingaccountId]
                : null,
            }}
          />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
  );
}

export function useBillingAccount() {
  return useOutletContext<ContextType>();
}
export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 billing account created</h3>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
