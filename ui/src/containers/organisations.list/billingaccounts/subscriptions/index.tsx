import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { V1Beta1Organization } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useContext, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { OrganizationsHeader } from "../../header";
import { getColumns } from "./columns";
import { AppContext } from "~/contexts/App";

export default function OrganisationBASubscriptions() {
  const { client } = useFrontier();
  const { plans } = useContext(AppContext);
  let { organisationId, billingaccountId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [subscriptions, setSubscriptions] = useState([]);

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
      const {
        // @ts-ignore
        data: { organization },
      } = await client?.frontierServiceGetOrganization(organisationId ?? "");
      setOrganisation(organization);
    }
    getOrganization();
  }, [client, organisationId]);

  useEffect(() => {
    async function getOrganizationSubscriptions() {
      const {
        // @ts-ignore
        data: { subscriptions },
      } = await client?.frontierServiceListSubscriptions(
        organisationId ?? "",
        billingaccountId ?? ""
      );
      setSubscriptions(subscriptions);
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
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 subsctription created</h3>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
