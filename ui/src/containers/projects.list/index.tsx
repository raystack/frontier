import { EmptyState, Flex, Table } from "@raystack/apsara";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import useSWR from "swr";
import { tableStyle } from "~/styles";
import { Project } from "~/types/project";
import { fetcher, reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { ProjectsHeader } from "./header";

type ContextType = { project: Project | null };
export default function ProjectList() {
  const { data, error } = useSWR("/v1beta1/admin/projects", fetcher);

  const { projects = [] } = data || { projects: [] };
  let { projectId } = useParams();

  const projectMapByName = reduceByKey(projects ?? [], "id");
  return (
    <Flex direction="row" css={{ height: "100%", width: "100%" }}>
      <Table
        css={tableStyle}
        columns={getColumns(projects)}
        data={projects ?? []}
        noDataChildren={noDataChildren}
      >
        <Table.TopContainer>
          <ProjectsHeader />
        </Table.TopContainer>
        <Table.DetailContainer
          css={{
            borderLeft: "1px solid $gray4",
            borderTop: "1px solid $gray4",
          }}
        >
          <Outlet
            context={{
              project: projectId ? projectMapByName[projectId] : null,
            }}
          />
        </Table.DetailContainer>
      </Table>
    </Flex>
  );
}

export function useProject() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState>
    <div className="svg-container"></div>
    <h3>0 project created</h3>
    <div className="pera">Try creating a new project.</div>
  </EmptyState>
);
