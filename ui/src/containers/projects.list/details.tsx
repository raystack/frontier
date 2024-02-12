import { Flex, Grid, Text } from "@raystack/apsara";
import { V1Beta1Project, V1Beta1User } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { ColumnDef } from "@tanstack/table-core";
import { useEffect, useState } from "react";
import { Link, NavLink, useParams } from "react-router-dom";
import PageHeader from "~/components/page-header";

type DetailsProps = {
  key: string;
  value: any;
};

export const userColumns: ColumnDef<V1Beta1User, any>[] = [
  {
    header: "Name",
    accessorKey: "name",
    cell: (info) => info.getValue(),
  },
  {
    header: "Email",
    accessorKey: "email",
    cell: (info) => info.getValue(),
  },
];
export default function ProjectDetails() {
  const { client } = useFrontier();
  let { projectId } = useParams();
  const [project, setProject] = useState<V1Beta1Project>();
  const [projectUsers, setProjectUsers] = useState([]);

  const pageHeader = {
    title: "Projects",
    breadcrumb: [
      {
        href: `/projects`,
        name: `Projects list`,
      },
      {
        href: `/projects/${project?.id}`,
        name: `${project?.name}`,
      },
    ],
  };

  useEffect(() => {
    async function getProject() {
      const {
        // @ts-ignore
        data: { project },
      } = await client?.frontierServiceGetProject(projectId ?? "");
      setProject(project);
    }
    getProject();
  }, [projectId]);

  useEffect(() => {
    async function getProjectUsers() {
      const {
        // @ts-ignore
        data: { users },
      } = await client?.frontierServiceListProjectUsers(projectId ?? "");
      setProjectUsers(users);
    }
    getProjectUsers();
  }, [projectId]);

  const detailList: DetailsProps[] = [
    {
      key: "Slug",
      value: project?.name,
    },
    {
      key: "Organization Id",
      value: project?.org_id,
    },
    {
      key: "Created At",
      value: new Date(project?.created_at as any).toLocaleString("en", {
        month: "long",
        day: "numeric",
        year: "numeric",
      }),
    },

    {
      key: "Users",
      value: (
        <Link to={`/projects/${projectId}/users`}>{projectUsers.length}</Link>
      ),
    },
  ];

  return (
    <Flex
      direction="column"
      gap="large"
      style={{
        width: "100%",
        height: "calc(100vh - 60px)",
        borderLeft: "1px solid var(--border-base)",
      }}
    >
      <PageHeader
        title={pageHeader.title}
        breadcrumb={pageHeader.breadcrumb}
        style={{ borderBottom: "1px solid var(--border-base)" }}
      >
        <NavLink
          to={`/projects/${projectId}/users`}
          style={{
            display: "flex",
            alignItems: "center",
            flexDirection: "row",
          }}
        >
          Users
        </NavLink>
      </PageHeader>
      <Flex direction="column" gap="large" style={{ padding: "0 24px" }}>
        {detailList.map((detailItem) => (
          <Grid columns={2} gap="small" key={detailItem.key}>
            <Text size={1}>{detailItem.key}</Text>
            <Text size={1}>{detailItem.value}</Text>
          </Grid>
        ))}
      </Flex>
    </Flex>
  );
}
