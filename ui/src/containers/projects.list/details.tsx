import { Dialog, Flex, Grid, Text } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { ColumnDef } from "@tanstack/table-core";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import DialogTable from "~/components/DialogTable";
import { DialogHeader } from "~/components/dialog/header";
import PageHeader from "~/components/page-header";
import { Project } from "~/types/project";
import { User } from "~/types/user";

type DetailsProps = {
  key: string;
  value: any;
};

export const userColumns: ColumnDef<User, any>[] = [
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
  const [project, setProject] = useState<Project>();
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
      } = await client?.frontierServiceListProjectUsers(project?.id ?? "");
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
      key: "Created At",
      value: new Date(project?.created_at as Date).toLocaleString("en", {
        month: "long",
        day: "numeric",
        year: "numeric",
      }),
    },
    {
      key: "Users",
      value: (
        <Dialog>
          <Dialog.Trigger>{projectUsers.length}</Dialog.Trigger>
          <Dialog.Content>
            <DialogTable
              columns={userColumns}
              data={projectUsers}
              header={<DialogHeader title="Organization users" />}
            />
          </Dialog.Content>
        </Dialog>
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
      />
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
