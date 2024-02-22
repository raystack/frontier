import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";

import { V1Beta1Organization } from "@raystack/frontier";
import { getColumns } from "./columns";
import { OrganizationsHeader } from "./header";

type ContextType = { organisation: V1Beta1Organization | null };
export default function OrganisationList() {
  const { client } = useFrontier();
  const [enabledOrganizations, setEnabledOrganizations] = useState<
    V1Beta1Organization[]
  >([]);
  const [disableddOrganizations, setDisabledOrganizations] = useState<
    V1Beta1Organization[]
  >([]);
  const [isOrganizationsLoading, setIsOrganizationsLoading] = useState(false);

  useEffect(() => {
    async function getOrganizations() {
      setIsOrganizationsLoading(true);
      try {
        const [orgResp, disabledOrgResp] = await Promise.all([
          client?.adminServiceListAllOrganizations(),
          client?.adminServiceListAllOrganizations({ state: "disabled" }),
        ]);
        if (orgResp?.data?.organizations) {
          setEnabledOrganizations(orgResp?.data?.organizations);
        }
        if (disabledOrgResp?.data?.organizations) {
          setDisabledOrganizations(disabledOrgResp?.data?.organizations);
        }
      } catch (err) {
        console.error(err);
      } finally {
        setIsOrganizationsLoading(false);
      }
    }
    getOrganizations();
  }, [client]);

  const organizations: V1Beta1Organization[] = isOrganizationsLoading
    ? [...new Array(5)].map((_, i) => ({
        name: i.toString(),
        title: "",
      }))
    : [...enabledOrganizations, ...disableddOrganizations];
  const tableStyle = organizations?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const columns = getColumns({
    isLoading: isOrganizationsLoading,
  });
  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={organizations ?? []}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <OrganizationsHeader />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet />
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
