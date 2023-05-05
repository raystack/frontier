import { EmptyState, Flex, Table } from "@odpf/apsara";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import useSWR from "swr";
import { tableStyle } from "~/styles";
import { Organisation } from "~/types/organisation";
import { fetcher, reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { OrganizationsHeader } from "./header";

type ContextType = { organisation: Organisation | null };
export default function OrganisationList() {
  const { data, error } = useSWR("/v1beta1/admin/organizations", fetcher);
  const { organizations = [] } = data || { organizations: [] };
  let { organisationId } = useParams();

  const organisationMapByName = reduceByKey(organizations ?? [], "id");
  return (
    <Flex direction="row" css={{ height: "100%", width: "100%" }}>
      <Table
        css={tableStyle}
        columns={getColumns(organizations)}
        data={organizations ?? []}
        noDataChildren={noDataChildren}
      >
        <Table.TopContainer>
          <OrganizationsHeader />
        </Table.TopContainer>
        <Table.DetailContainer
          css={{
            borderLeft: "1px solid $gray4",
            borderTop: "1px solid $gray4",
          }}
        >
          {organisationId && (
            <Outlet
              context={{
                organisation: organisationMapByName[organisationId],
              }}
            />
          )}
        </Table.DetailContainer>
      </Table>
      {!organisationId && <Outlet />}
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
