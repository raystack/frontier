import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useContext, useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import {
  V1Beta1Organization,
  V1Beta1ServiceUser,
  V1Beta1User,
} from "@raystack/frontier";
import { getColumns } from "./columns";
import { OrganizationsServiceUsersHeader } from "./header";
import { AppContext } from "~/contexts/App";

type ContextType = { user: V1Beta1User | null };
export default function OrganisationServiceUsers() {
  const { client } = useFrontier();
  let { organisationId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [serviceusers, setOrgServiceUsers] = useState<V1Beta1ServiceUser[]>([]);
  const { platformUsers } = useContext(AppContext);

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
      const resp = await client?.frontierServiceGetOrganization(
        organisationId ?? ""
      );
      const organization = resp?.data?.organization;
      setOrganisation(organization);
    }
    getOrganization();
  }, [client, organisationId]);

  useEffect(() => {
    async function getOrganizationUser() {
      const resp = await client?.frontierServiceListServiceUsers({
        org_id: organisationId ?? "",
      });
      const serviceusers = resp?.data?.serviceusers || [];
      setOrgServiceUsers(serviceusers);
    }
    getOrganizationUser();
  }, [client, organisationId]);

  const tableStyle = serviceusers?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const columns = getColumns({
    orgId: organisationId || "",
    platformUsers: platformUsers?.serviceusers || [],
  });

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
