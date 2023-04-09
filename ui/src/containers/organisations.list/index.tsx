import { EmptyState, Flex, Table } from "@odpf/apsara";
import { Outlet, useOutletContext } from "react-router-dom";
import useSWR from "swr";
import { tableStyle } from "~/styles";
import { Organisation } from "~/types/organisation";
import { fetcher, reduceByKey } from "~/utils/helper";
import { columns } from "./columns";
import { OrganizationsHeader } from "./header";

type ContextType = { organizationMapByName: Record<string, Organisation> | {} };
export default function OrganisationList() {
  const { data, error } = useSWR("/admin/v1beta1/organizations", fetcher);
  const { organizations = [] } = data || { organizations: [] };

  const organizationMapByName = reduceByKey(organizations ?? [], "id");
  return (
    <Flex direction="row" css={{ height: "100%", width: "100%" }}>
      <Table
        css={tableStyle}
        columns={columns}
        data={organizations ?? []}
        noDataChildren={noDataChildren}
      >
        <Table.TopContainer>
          <OrganizationsHeader />
        </Table.TopContainer>
      </Table>
      <Outlet context={{ organizationMapByName }} />
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
