import { DataTable, EmptyState } from "@raystack/apsara";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import useSWR from "swr";
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
    <DataTable
      data={projects ?? []}
      // @ts-ignore
      columns={getColumns(projects)}
      emptyState={noDataChildren}
      style={{ width: "100%" }}
    >
      <DataTable.Toolbar>
        <ProjectsHeader />
        <DataTable.FilterChips style={{ paddingTop: "16px" }} />
      </DataTable.Toolbar>
      <DataTable.DetailContainer>
        <Outlet
          context={{
            project: projectId ? projectMapByName[projectId] : null,
          }}
        />
      </DataTable.DetailContainer>
    </DataTable>
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
