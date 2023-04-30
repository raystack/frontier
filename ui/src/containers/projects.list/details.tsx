import { Dialog, Flex, Grid, Text } from "@odpf/apsara";
import { ColumnDef } from "@tanstack/table-core";
import useSWR from "swr";
import { DialogHeader } from "~/components/dialog/header";
import DialogTable from "~/components/DialogTable";
import { User } from "~/types/user";
import { fetcher } from "~/utils/helper";
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
  const { project } = useProject();
  const { data: usersData } = useSWR(
    `/admin/v1beta1/organizations/${project?.id}/users`,
    fetcher
  );
  const { users = [] } = usersData || { users: [] };

  const detailList: DetailsProps[] = [
    {
      key: "Slug",
      value: project?.slug,
    },
    {
      key: "Created At",
      value: new Date(project?.createdAt as Date).toLocaleString("en", {
        month: "long",
        day: "numeric",
        year: "numeric",
      }),
    },
    {
      key: "Users",
      value: (
        <Dialog>
          <Dialog.Trigger style={css.button}>{users.length}</Dialog.Trigger>
          <Dialog.Content css={{ padding: 0, button: { my: "$2" } }}>
            <DialogTable
              columns={userColumns}
              data={users}
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
      css={{
        width: "320px",
        height: "100%",
        padding: "$4",
      }}
    >
      <Text css={{ fontSize: "14px" }}>{project?.name}</Text>
      <Flex direction="column">
        {detailList.map((detailItem) => (
          <Grid
            columns="2"
            css={{ width: "100%", paddingTop: "$4" }}
            key={detailItem.key}
          >
            <Text
              size={1}
              css={{
                color: "$gray11",
                ...css.row,
              }}
            >
              {detailItem.key}
            </Text>
            <Text size={1} css={css.row}>
              {detailItem.value}
            </Text>
          </Grid>
        ))}
      </Flex>
    </Flex>
  );
}

const css = {
  row: { height: "32px", display: "flex", alignItems: "center" },
  button: {
    background: "none",
    color: "inherit",
    border: "none",
    padding: "0",
    cursor: "pointer",
    display: "flex",
    alignItems: "center",
    height: "100%",
    width: "100%",
  },
};
