import { DataTable } from "@raystack/apsara";
import { EmptyState, Flex } from "@raystack/apsara/v1";
import { V1Beta1BillingAccount, V1Beta1Organization } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { reduceByKey } from "~/utils/helper";
import { OrganizationsHeader } from "../header";
import { getColumns } from "./columns";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

type ContextType = { billingaccount: V1Beta1BillingAccount | null };
export default function OrganisationBillingAccounts() {
  const { client } = useFrontier();
  let { organisationId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [billingAccounts, setBillingAccounts] = useState<
    V1Beta1BillingAccount[]
  >([]);

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
      try {
        const res = await client?.frontierServiceGetOrganization(
          organisationId ?? ""
        );
        const organization = res?.data?.organization;
        setOrganisation(organization);
      } catch (error) {
        console.error(error);
      }
    }
    getOrganization();
  }, [client, organisationId]);

  useEffect(() => {
    async function getOrganizationBillingAccounts() {
      try {
        const res = await client?.frontierServiceListBillingAccounts(
          organisationId ?? ""
        );
        const billing_accounts = res?.data?.billing_accounts ?? [];
        setBillingAccounts(billing_accounts);
      } catch (error) {
        console.error(error);
      }
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
  <EmptyState icon={<ExclamationTriangleIcon />} heading="No billing account" />
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
