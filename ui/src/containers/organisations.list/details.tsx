import { Button, Dialog, Flex, Grid, Text } from "@raystack/apsara";
import { ColumnDef } from "@tanstack/table-core";
import useSWR from "swr";
import DialogTable from "~/components/DialogTable";
import { DialogHeader } from "~/components/dialog/header";
import { User } from "~/types/user";
import { fetcher } from "~/utils/helper";
import { useOrganisation } from ".";

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
export const projectColumns: ColumnDef<User, any>[] = [
  {
    header: "Name",
    accessorKey: "name",
    cell: (info) => info.getValue(),
  },
  {
    header: "Slug",
    accessorKey: "slug",
    cell: (info) => info.getValue(),
  },
];

export default function OrganisationDetails() {
  const { organisation } = useOrganisation();
  console.log(organisation);
  const { data: usersData } = useSWR(
    `/v1beta1/organizations/${organisation?.id}/users`,
    fetcher
  );
  const { data: projectsData } = useSWR("/v1beta1/admin/projects", fetcher);
  const { users = [] } = usersData || { users: [] };
  const { projects = [] } = projectsData || { projects: [] };

  const detailList: DetailsProps[] = [
    {
      key: "Name",
      value: organisation?.name,
    },
    {
      key: "Created At",
      value: new Date(organisation?.created_at as Date).toLocaleString("en", {
        month: "long",
        day: "numeric",
        year: "numeric",
      }),
    },
    {
      key: "Users",
      value: (
        <Dialog>
          <Dialog.Trigger asChild>
            <Button>{users.length}</Button>
          </Dialog.Trigger>
          <Dialog.Content>
            <DialogTable
              columns={userColumns}
              data={users}
              header={<DialogHeader title="Organization users" />}
            />
          </Dialog.Content>
        </Dialog>
      ),
    },
    {
      key: "Projects",
      value: (
        <Dialog>
          <Dialog.Trigger asChild>
            <Button>{projects.length}</Button>
          </Dialog.Trigger>
          <Dialog.Content>
            <DialogTable
              columns={projectColumns}
              data={projects}
              header={<DialogHeader title="Organization project" />}
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
      <Text size={4}>{organisation?.name}</Text>
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
