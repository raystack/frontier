import { DataTable, EmptyState, Flex } from "@raystack/apsara";
import { V1Beta1Project } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useContext, useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { AppContext } from "~/contexts/App";
import { ProjectsHeader } from "./header";

type ContextType = { project: V1Beta1Project | null };
export default function ProjectList() {
  const { client } = useFrontier();
  const { orgMap } = useContext(AppContext);
  const [projects, setProjects] = useState<V1Beta1Project[]>([]);
  const [isProjectsLoading, setIsProjectsLoading] = useState(false);

  useEffect(() => {
    async function getProjects() {
      setIsProjectsLoading(true);
      try {
        const {
          // @ts-ignore
          data: { projects },
        } = await client?.adminServiceListProjects();
        setProjects(projects);
      } catch (err) {
        console.error(err);
      } finally {
        setIsProjectsLoading(false);
      }
    }
    getProjects();
  }, [client]);

  let { projectId } = useParams();
  const projectMapByName = reduceByKey(projects ?? [], "id");

  const tableStyle = projects?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  const projectList = isProjectsLoading
    ? [...new Array(5)].map((_, i) => ({
        name: i.toString(),
        title: "",
      }))
    : projects;

  const columns = getColumns({
    orgMap,
  });

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={projectList}
        // @ts-ignore
        columns={columns}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
        isLoading={isProjectsLoading}
      >
        <DataTable.Toolbar>
          <ProjectsHeader />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
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
