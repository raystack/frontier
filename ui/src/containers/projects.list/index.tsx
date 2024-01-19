import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { Project } from "~/types/project";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { ProjectsHeader } from "./header";

type ContextType = { project: Project | null };
export default function ProjectList() {
  const { client } = useFrontier();
  const [projects, setProjects] = useState([]);

  useEffect(() => {
    async function getProjects() {
      const {
        // @ts-ignore
        data: { projects },
      } = await client?.adminServiceListProjects();
      setProjects(projects);
    }
    getProjects();
  }, []);

  let { projectId } = useParams();
  const projectMapByName = reduceByKey(projects ?? [], "id");

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
