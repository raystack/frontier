import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { useOutletContext, useParams } from "react-router-dom";
import { Organisation } from "~/types/organisation";
import { User } from "~/types/user";
import { OrganizationsHeader } from "../header";
import { getColumns } from "./columns";

type ContextType = { user: User | null };
export default function OrganisationServiceUsers() {
  const { client } = useFrontier();
  let { organisationId } = useParams();
  const [organisation, setOrganisation] = useState<Organisation>();
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
  }, [organisationId]);

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
  }, [organisationId]);

  const tableStyle = serviceusers?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={serviceusers ?? []}
        // @ts-ignore
        columns={getColumns(serviceusers)}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <OrganizationsHeader header={pageHeader} />
          <DataTable.FilterChips style={{ paddingTop: "16px" }} />
        </DataTable.Toolbar>
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
    <h3>0 service user created</h3>
    <div className="pera">Try creating a new service user.</div>
  </EmptyState>
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
