import { Dialog, Flex, Grid, Text } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { ColumnDef } from "@tanstack/table-core";
import { useEffect, useState } from "react";
import DialogTable from "~/components/DialogTable";
import { DialogHeader } from "~/components/dialog/header";
import { User } from "~/types/user";
import { useProject } from ".";

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
  const { project } = useProject();
  const [projectUsers, setProjectUsers] = useState([]);

  useEffect(() => {
    async function getOrganizations() {
      const {
        // @ts-ignore
        data: { users },
      } = await client?.frontierServiceListProjectUsers(project?.id ?? "");
      setProjectUsers(users);
    }
    getOrganizations();
  }, [project?.id]);

  const detailList: DetailsProps[] = [
    {
      key: "Slug",
      value: project?.slug,
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
        width: "320px",
        height: "calc(100vh - 60px)",
        borderLeft: "1px solid var(--border-base)",
        padding: "var(--pd-16)",
      }}
    >
      <Text size={4}>{project?.name}</Text>
      <Flex direction="column" gap="large">
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
