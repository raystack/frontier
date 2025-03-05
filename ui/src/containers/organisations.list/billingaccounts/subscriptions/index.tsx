import { DataTable } from "@raystack/apsara";
import { EmptyState, Flex } from "@raystack/apsara/v1";

import { V1Beta1Organization, V1Beta1Subscription } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useContext, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { OrganizationsHeader } from "../../header";
import { getColumns } from "./columns";
import { AppContext } from "~/contexts/App";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

export default function OrganisationBASubscriptions() {
  const { client } = useFrontier();
  const { plans } = useContext(AppContext);
  let { organisationId, billingaccountId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [subscriptions, setSubscriptions] = useState<V1Beta1Subscription[]>([]);

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
        name: `Subscriptions`,
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
    async function getOrganizationSubscriptions() {
      try {
        const res = await client?.frontierServiceListSubscriptions(
          organisationId ?? "",
          billingaccountId ?? ""
        );
        const subscriptions = res?.data?.subscriptions ?? [];
        setSubscriptions(subscriptions);
      } catch (error) {
        console.error(error);
      }
    }
    getOrganizationSubscriptions();
  }, [billingaccountId, client, organisationId]);

  const tableStyle = subscriptions?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const columns = getColumns({ subscriptions, plans });
  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={subscriptions ?? []}
        // @ts-ignore
        columns={columns}
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
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="0 subscriptions created"
  />
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
