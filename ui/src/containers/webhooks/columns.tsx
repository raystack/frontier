import { DotsVerticalIcon, TrashIcon, UpdateIcon } from "@radix-ui/react-icons";
import { ApsaraColumnDef, DropdownMenu, Flex } from "@raystack/apsara";
import { V1Beta1Webhook } from "@raystack/frontier";

interface getColumnsOptions {
  openEditPage: (id: string) => void;
}

export const getColumns: (
  opt: getColumnsOptions
) => ApsaraColumnDef<V1Beta1Webhook>[] = ({ openEditPage }) => {
  return [
    {
      header: "Description",
      accessorKey: "description",
      filterVariant: "text",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "State",
      accessorKey: "state",
      filterVariant: "text",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "URL",
      accessorKey: "url",
      cell: (info) => info.getValue() || "-",
    },
    {
      header: "Created at",
      accessorKey: "created_at",
      cell: (info) =>
        new Date(info.getValue() as Date).toLocaleString("en", {
          month: "long",
          day: "numeric",
          year: "numeric",
        }),
    },
    {
      header: "Action",
      accessorKey: "id",
      cell: ({ getValue }) => (
        <DropdownMenu style={{ padding: "0 !important" }}>
          <DropdownMenu.Trigger asChild style={{ cursor: "pointer" }}>
            <DotsVerticalIcon />
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end">
            <DropdownMenu.Group style={{ padding: 0 }}>
              <DropdownMenu.Item style={{ padding: 0 }}>
                <Flex
                  style={{ padding: "12px" }}
                  gap={"small"}
                  data-test-id="admin-ui-webhook-update-btn"
                  onClick={() => openEditPage(getValue())}
                >
                  <UpdateIcon />
                  Update
                </Flex>
              </DropdownMenu.Item>
              <DropdownMenu.Item style={{ padding: 0 }} disabled>
                <Flex
                  style={{ padding: "12px" }}
                  gap={"small"}
                  data-test-id="admin-ui-webhook-delete-btn"
                >
                  <TrashIcon />
                  Delete
                </Flex>
              </DropdownMenu.Item>
            </DropdownMenu.Group>
          </DropdownMenu.Content>
        </DropdownMenu>
      ),
    },
  ];
};
