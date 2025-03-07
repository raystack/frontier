import { DataTable } from "@raystack/apsara";
import { EmptyState, Flex } from "@raystack/apsara/v1";
import { V1Beta1Project } from "@raystack/frontier";
import { useContext, useEffect, useState } from "react";
import { Outlet, useOutletContext, useParams } from "react-router-dom";
import { reduceByKey } from "~/utils/helper";
import { getColumns } from "./columns";
import { AppContext } from "~/contexts/App";
import { ProjectsHeader } from "./header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";

type ContextType = { project: V1Beta1Project | null };
export default function ProjectList() {
  const { orgMap } = useContext(AppContext);
  const [projects, setProjects] = useState<V1Beta1Project[]>([]);
  const [isProjectsLoading, setIsProjectsLoading] = useState(false);

  useEffect(() => {
    async function getProjects() {
      setIsProjectsLoading(true);
      try {
        const resp = (await api?.adminServiceListProjects()) || {};
        const newProjects = resp?.data?.projects ?? [];
        setProjects(newProjects);
      } catch (err) {
        console.error(err);
      } finally {
        setIsProjectsLoading(false);
      }
    }
    getProjects();
  }, []);

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

const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="No projects found"
    subHeading="There are no projects in this organization."
  />
);
