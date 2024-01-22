import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useMemo, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { Organisation } from "~/types/organisation";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { OrgStates, OrganizationsHeader } from "./header";

type ContextType = { organisation: Organisation | null };
export default function OrganisationList() {
  const { client } = useFrontier();
  const [orgState, setOrgState] = useState<OrgStates>("enabled");
  const [organizations, setOrganizations] = useState([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    async function getOrganizations() {
      setIsLoading(true);
      try {
        const {
          // @ts-ignore
          data: { organizations },
        } = await client?.adminServiceListAllOrganizations({
          state: orgState,
        });
        setOrganizations(organizations);
      } catch (err) {
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    }
    getOrganizations();
  }, [orgState]);

  let { organisationId } = useParams();
  const organisationMapByName = reduceByKey(organizations ?? [], "id");

  const tableStyle = organizations?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const columns = getColumns({ organizations, isLoading });

  const onStateFilterChange = (state: OrgStates) => {
    setOrgState(state);
  };

  const orgList = isLoading
    ? [...new Array(10)].map((_, i) => {
        id: i;
      })
    : organizations ?? [];

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        // @ts-ignore
        data={orgList}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <OrganizationsHeader onStateFilterChange={onStateFilterChange} on />
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
