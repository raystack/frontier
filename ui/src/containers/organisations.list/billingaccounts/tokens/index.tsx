import { DataTable, EmptyState, Flex, Text } from "@raystack/apsara";
import {
  V1Beta1BillingAccount,
  V1Beta1BillingTransaction,
  V1Beta1Organization,
} from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { OrganizationsTokenHeader } from "./header";
import { toast } from "sonner";
import { getColumns } from "./columns";
import { useTokens } from "./useTokens";
import Skeleton from "react-loading-skeleton";

type ContextType = { billingaccount: V1Beta1BillingAccount | null };
export default function OrganisationTokens() {
  const { client } = useFrontier();
  let { organisationId, billingaccountId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [transactionsList, setTransactionsList] = useState<
    V1Beta1BillingTransaction[]
  >([]);
  const [isTransactionsListLoading, setIsTransactionsListLoading] =
    useState(false);

  const { tokenBalance, isTokensLoading } = useTokens({
    organisationId,
    billingaccountId,
  });
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
        href: ``,
        name: `Tokens`,
      },
    ],
  };

  useEffect(() => {
    async function getOrganization(orgId: string) {
      try {
        const res = await client?.frontierServiceGetOrganization(orgId)
        const organization = res?.data?.organization
        setOrganisation(organization);
      } catch (error) {
        console.error(error)
      }
    }

    async function getTransactions(orgId: string, billingAccountId: string) {
      try {
        setIsTransactionsListLoading(true);
        const resp = await client?.frontierServiceListBillingTransactions(
          orgId,
          billingAccountId,
          {
            expand: ["user"],
          }
        );
        const txns = resp?.data?.transactions || [];
        setTransactionsList(txns);
      } catch (err: any) {
        console.error(err);
        toast.error("Unable to fetch transactions");
      } finally {
        setIsTransactionsListLoading(false);
      }
    }

    if (organisationId && billingaccountId) {
      getOrganization(organisationId);
      getTransactions(organisationId, billingaccountId);
    }
  }, [client, organisationId, billingaccountId]);

  const tableStyle = transactionsList?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const columns = getColumns({ isLoading: isTransactionsListLoading });

  const isLoading = isTokensLoading || isTransactionsListLoading;

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={transactionsList ?? []}
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <OrganizationsTokenHeader header={pageHeader} />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
          <Flex
            style={{
              padding: "12px 24px",
              borderTop: "1px solid var(--border-base)",
            }}
          >
            <Text size={3} weight={500}>
              Balance:
            </Text>
            {isLoading ? <Skeleton /> : <Text size={3}>{tokenBalance}</Text>}
          </Flex>
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet />
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
    <h3>No token transactions</h3>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
