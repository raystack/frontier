import { Button, Dialog, Flex, Grid, Text } from "@raystack/apsara";
import { useFrontier } from "@raystack/frontier/react";
import { ColumnDef } from "@tanstack/table-core";
import { useEffect, useState } from "react";
import DialogTable from "~/components/DialogTable";
import { DialogHeader } from "~/components/dialog/header";
import { User } from "~/types/user";
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
  const { client } = useFrontier();
  const { organisation } = useOrganisation();
  const [orgUsers, setOrgUsers] = useState([]);
  const [orgProjects, setOrgProjects] = useState([]);

  useEffect(() => {
    async function getOrganizationUser() {
      const {
        // @ts-ignore
        data: { users },
      } = await client?.frontierServiceListOrganizationUsers(
        organisation?.id ?? ""
      );
      setOrgUsers(users);
    }
    getOrganizationUser();
  }, [organisation?.id]);

  useEffect(() => {
    async function getOrganizationProjects() {
      const {
        // @ts-ignore
        data: { projects },
      } = await client?.frontierServiceListOrganizationProjects(
        organisation?.id ?? ""
      );
      setOrgProjects(projects);
    }
    getOrganizationProjects();
  }, [organisation?.id ?? ""]);

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
            <Button>{orgUsers.length}</Button>
          </Dialog.Trigger>
          <Dialog.Content>
            <DialogTable
              columns={userColumns}
              data={orgUsers}
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
            <Button>{orgProjects.length}</Button>
          </Dialog.Trigger>
          <Dialog.Content>
            <DialogTable
              columns={projectColumns}
              data={orgProjects}
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
