import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import { V1Beta1Organization, V1Beta1User } from "@raystack/frontier";
import { OrganizationsHeader } from "../header";
import { getColumns } from "./columns";
import { reduceByKey } from "~/utils/helper";

type ContextType = { user: V1Beta1User | null };
export default function OrganisationUsers() {
  const { client } = useFrontier();
  let { organisationId, userId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [users, setOrgUsers] = useState<V1Beta1User[]>([]);

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
        name: `Organizations Users`,
      },
    ],
  };

  useEffect(() => {
    async function getOrganization(orgId: string) {
      const {
        // @ts-ignore
        data: { organization },
      } = await client?.frontierServiceGetOrganization(orgId);
      setOrganisation(organization);
    }
    async function getOrganizationUser(orgId: string) {
      const resp = await client?.frontierServiceListOrganizationUsers(orgId, {
        with_roles: true,
      });
      const userList = resp?.data?.users || [];
      setOrgUsers(userList);
    }
    if (organisationId) {
      getOrganization(organisationId);
      getOrganizationUser(organisationId);
    }
  }, [client, organisationId]);

  const tableStyle = users?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const userMapById = reduceByKey(users ?? [], "id");

  const columns = getColumns({ users, orgId: organisationId || "" });

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={users ?? []}
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
        <DataTable.DetailContainer>
          <Outlet
            context={{
              user: userId ? userMapById[userId] : null,
            }}
          />
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
    <h3>No users created</h3>
    <div className="pera">Try creating a new user.</div>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
