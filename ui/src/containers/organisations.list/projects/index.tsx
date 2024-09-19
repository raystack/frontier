import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { V1Beta1Organization, V1Beta1Project, V1Beta1User } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { useOutletContext, useParams } from "react-router-dom";
import { OrganizationsHeader } from "../header";
import { getColumns } from "./columns";

type ContextType = { user: V1Beta1User | null };
export default function OrganisationProjects() {
  const { client } = useFrontier();
  let { organisationId } = useParams();
  const [organisation, setOrganisation] = useState<V1Beta1Organization>();
  const [projects, setOrgProjects] = useState<V1Beta1Project[]>([]);

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
        name: `Organizations Projects`,
      },
    ],
  };

  useEffect(() => {
    async function getOrganization() {
      try {
        const res = await client?.frontierServiceGetOrganization(organisationId ?? "")
        const organization = res?.data?.organization
        setOrganisation(organization);
      } catch (error) {
        console.error(error)
      }
    }
    getOrganization();
  }, [organisationId]);

  useEffect(() => {
    async function getOrganizationProjects() {
      try {
        const res = await client?.frontierServiceListOrganizationProjects(organisationId ?? "")
        const projects = res?.data?.projects ?? []
        setOrgProjects(projects);
      } catch (error) {
        console.error(error)
      }
    }
    getOrganizationProjects();
  }, [organisationId ?? ""]);

  const tableStyle = projects?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={projects ?? []}
        // @ts-ignore
        columns={getColumns(projects)}
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
