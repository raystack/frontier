import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import { V1Beta1Organization, V1Beta1User } from "@raystack/frontier";
import { OrganizationsHeader } from "../header";
import { getColumns } from "./columns";
import { OrganizationsServiceUsersHeader } from "./header";

type ContextType = { user: V1Beta1User | null };
export default function OrganisationServiceUsers() {
  const { client } = useFrontier();
  let { organisationId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [serviceusers, setOrgServiceUsers] = useState([]);

  const pageHeader = {
    title: "Organizations",
    breadcrumb: [
      {
        href: `/organisations`,
        name: `Organizations list`,
      },
      {
        href: `/organisations/${organisationId}`,
        name: `${organisation?.name}`,
      },
      {
        href: ``,
        name: `Organizations Service Users`,
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
    async function getOrganizationUser() {
      const {
        // @ts-ignore
        data: { serviceusers },
      } = await client?.frontierServiceListServiceUsers({
        org_id: organisationId ?? "",
      });
      setOrgServiceUsers(serviceusers);
    }
    getOrganizationUser();
  }, [client, organisationId]);

  const tableStyle = serviceusers?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const columns = getColumns({ orgId: organisationId || "" });

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={serviceusers ?? []}
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <OrganizationsServiceUsersHeader
            header={pageHeader}
            orgId={organisationId}
          />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
  );
}

export function useUser() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>No service users created</h3>
    <div className="pera">Try creating a new service user.</div>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
