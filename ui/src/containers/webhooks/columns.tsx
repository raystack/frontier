import { DotsVerticalIcon, TrashIcon, UpdateIcon } from "@radix-ui/react-icons";
import type { V1Beta1Webhook } from "@raystack/frontier";
import { DropdownMenu, Flex, type DataTableColumnDef } from "@raystack/apsara";
import styles from "./webhooks.module.css";

interface getColumnsOptions {
  openEditPage: (id: string) => void;
}

export const getColumns: (
  opt: getColumnsOptions,
) => DataTableColumnDef<V1Beta1Webhook, unknown>[] = ({ openEditPage }) => {
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
      classNames: { cell: styles.stateColumn, header: styles.stateColumn },
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
      classNames: { cell: styles.dateColumn, header: styles.dateColumn },
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
      classNames: { cell: styles.actionColumn, header: styles.actionColumn },
      cell: ({ getValue }) => (
        // @ts-ignore
        <DropdownMenu style={{ padding: "0 !important" }}>
          <DropdownMenu.Trigger asChild style={{ cursor: "pointer" }}>
            <DotsVerticalIcon />
          </DropdownMenu.Trigger>
          <DropdownMenu.Content>
            <DropdownMenu.Group style={{ padding: 0 }}>
              <DropdownMenu.Item style={{ padding: 0 }}>
                <Flex
                  style={{ padding: "12px" }}
                  gap={"small"}
                  data-test-id="admin-ui-webhook-update-btn"
                  onClick={() => openEditPage(getValue() as string)}
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
