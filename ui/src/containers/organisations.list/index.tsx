import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { Organisation } from "~/types/organisation";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { OrganizationsHeader } from "./header";

type ContextType = { organisation: Organisation | null };
export default function OrganisationList() {
  const { client } = useFrontier();
  const [organizations, setOrganizations] = useState([]);

  useEffect(() => {
    async function getOrganizations() {
      const {
        // @ts-ignore
        data: { organizations },
      } = await client?.adminServiceListAllOrganizations();
      setOrganizations(organizations);
    }
    getOrganizations();
  }, []);

  let { organisationId } = useParams();
  const organisationMapByName = reduceByKey(organizations ?? [], "id");

  const tableStyle = organizations?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={organizations ?? []}
        // @ts-ignore
        columns={getColumns(organizations)}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <OrganizationsHeader />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet
            context={{
              organisation: organisationId
                ? organisationMapByName[organisationId]
                : null,
            }}
          />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
  );
}

export function useOrganisation() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 organisation created</h3>
    <div className="pera">Try creating a new organisation.</div>
  </EmptyState>
);
