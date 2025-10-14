import { DotsVerticalIcon, TrashIcon, UpdateIcon } from "@radix-ui/react-icons";
import {
  DropdownMenu,
  Flex,
  Text,
  type DataTableColumnDef,
} from "@raystack/apsara";
import styles from "./webhooks.module.css";
import { type Webhook } from "@raystack/proton/frontier";
import {
  timestampToDate,
  isNullTimestamp,
  TimeStamp,
} from "~/utils/connect-timestamp";
import dayjs from "dayjs";
import { useState } from "react";
import type { useMutation } from "@connectrpc/connect-query";
import { DeleteWebhookDialog } from "./delete";

interface getColumnsOptions {
  openEditPage: (id: string) => void;
  deleteWebhookMutation: ReturnType<typeof useMutation>;
  enableDelete: boolean;
}

export const getColumns: (
  opt: getColumnsOptions,
) => DataTableColumnDef<Webhook, unknown>[] = ({ openEditPage, deleteWebhookMutation, enableDelete }) => {
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
      accessorKey: "createdAt",
      classNames: { cell: styles.dateColumn, header: styles.dateColumn },
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        const date = isNullTimestamp(value)
          ? "-"
          : dayjs(timestampToDate(value)).format("YYYY-MM-DD");
        return <Text>{date}</Text>;
      },
    },
    {
      header: "Action",
      accessorKey: "id",
      classNames: { cell: styles.actionColumn, header: styles.actionColumn },
      cell: ({ getValue, row }) => {
        const ActionCell = () => {
          const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
          const webhookId = getValue() as string;
          const webhook = row.original;

          return (
            <>
              {/* @ts-ignore */}
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
                        onClick={() => openEditPage(webhookId)}
                      >
                        <UpdateIcon />
                        Update
                      </Flex>
                    </DropdownMenu.Item>
                    <DropdownMenu.Item style={{ padding: 0 }} disabled={!enableDelete}>
                      <Flex
                        className={styles.deleteMenuItem}
                        gap={"small"}
                        data-test-id="admin-ui-webhook-delete-btn"
                        onClick={() => enableDelete && setIsDeleteDialogOpen(true)}
                      >
                        <TrashIcon />
                        Delete
                      </Flex>
                    </DropdownMenu.Item>
                  </DropdownMenu.Group>
                </DropdownMenu.Content>
              </DropdownMenu>

              <DeleteWebhookDialog
                isOpen={isDeleteDialogOpen}
                onOpenChange={setIsDeleteDialogOpen}
                webhookId={webhookId}
                webhookDescription={webhook.description}
                deleteWebhookMutation={deleteWebhookMutation}
              />
            </>
          );
        };

        return <ActionCell />;
      },
    },
  ];
};
