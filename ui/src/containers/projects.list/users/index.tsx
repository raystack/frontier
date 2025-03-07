import { DataTable } from "@raystack/apsara";
import { EmptyState, Flex } from "@raystack/apsara/v1";
import { useEffect, useState } from "react";
import { useOutletContext, useParams } from "react-router-dom";

import { V1Beta1Project, V1Beta1User } from "@raystack/frontier";
import { getColumns } from "./columns";
import { ProjectsHeader } from "../header";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { api } from "~/api";

type ContextType = { user: V1Beta1User | null };
export default function ProjectUsers() {
  let { projectId } = useParams();
  const [project, setProject] = useState<V1Beta1Project>();
  const [users, setProjectUsers] = useState<V1Beta1User[]>([]);

  const pageHeader = {
    title: "Projects",
    breadcrumb: [
      {
        href: `/projects`,
        name: `Projects list`,
      },
      {
        href: `/projects/${projectId}`,
        name: `${project?.name}`,
      },
      {
        href: ``,
        name: `Projects Users`,
      },
    ],
  };

  useEffect(() => {
    async function getProject() {
      const {
        // @ts-ignore
        data: { project },
      } = (await api?.frontierServiceGetProject(projectId ?? "")) || {};
      setProject(project);
    }
    getProject();
  }, [projectId]);

  useEffect(() => {
    async function getProjectUser() {
      const resp =
        (await api?.frontierServiceListProjectUsers(projectId ?? "")) || {};
      const newUsers = resp?.data?.users ?? [];
      setProjectUsers(newUsers);
    }
    getProjectUser();
  }, [projectId]);

  const tableStyle = users?.length
    ? { width: "100%" }
    : { width: "100%", height: "100%" };

  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={users ?? []}
        // @ts-ignore
        columns={getColumns(users)}
        emptyState={noDataChildren}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <ProjectsHeader header={pageHeader} />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
}

export function useUser() {
  return useOutletContext<ContextType>();
}

export const noDataChildren = (
  <EmptyState
    icon={<ExclamationTriangleIcon />}
    heading="No users created"
    subHeading="Try creating a new user."
  />
);

export const TableDetailContainer = ({ children }: any) => (
  <div>{children}</div>
);
