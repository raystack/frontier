import { DotsVerticalIcon, TrashIcon, UpdateIcon } from "@radix-ui/react-icons";
import {
  DropdownMenu,
  Flex,
  Text,
  type DataTableColumnDef,
  Dialog,
  Button,
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
import { toast } from "sonner";
import type { useMutation } from "@connectrpc/connect-query";

interface getColumnsOptions {
  openEditPage: (id: string) => void;
  deleteWebhookMutation: ReturnType<typeof useMutation>;
}

interface DeleteConfirmDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isLoading: boolean;
  webhookDescription?: string;
}

function DeleteConfirmDialog({
  isOpen,
  onOpenChange,
  onConfirm,
  isLoading,
  webhookDescription,
}: DeleteConfirmDialogProps) {
  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <Dialog.Content
        style={{
          maxWidth: "25vw",
          width: "100%",
        }}
      >
        <Flex direction="column" gap="large" style={{ padding: "24px" }}>
          <Flex direction="column" gap="medium">
            <Text size={5} weight={500}>
              Delete Webhook
            </Text>
            <Text>
              Are you sure you want to delete this webhook
              {webhookDescription ? ` "${webhookDescription}"` : ""}? This action
              cannot be undone.
            </Text>
          </Flex>

          <Flex justify="end" gap={5}>
            <Button
              variant="outline"
              color="neutral"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
              data-test-id="admin-ui-cancel-delete-webhook"
            >
              Cancel
            </Button>
            <Button
              variant="solid"
              color="danger"
              onClick={onConfirm}
              loading={isLoading}
              loaderText="Deleting..."
              data-test-id="admin-ui-confirm-delete-webhook"
            >
              Delete
            </Button>
          </Flex>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}

export const getColumns: (
  opt: getColumnsOptions,
) => DataTableColumnDef<Webhook, unknown>[] = ({ openEditPage, deleteWebhookMutation }) => {
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

          const handleDelete = async () => {
            try {
              await deleteWebhookMutation.mutateAsync({ id: webhookId });
              toast.success("Webhook deleted");
              setIsDeleteDialogOpen(false);
            } catch (err) {
              console.error("Failed to delete webhook:", err);
              toast.error("Failed to delete webhook");
            }
          };

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
                    <DropdownMenu.Item style={{ padding: 0 }}>
                      <Flex
                        className={styles.deleteMenuItem}
                        gap={"small"}
                        data-test-id="admin-ui-webhook-delete-btn"
                        onClick={() => setIsDeleteDialogOpen(true)}
                      >
                        <TrashIcon />
                        Delete
                      </Flex>
                    </DropdownMenu.Item>
                  </DropdownMenu.Group>
                </DropdownMenu.Content>
              </DropdownMenu>

              <DeleteConfirmDialog
                isOpen={isDeleteDialogOpen}
                onOpenChange={setIsDeleteDialogOpen}
                onConfirm={handleDelete}
                isLoading={deleteWebhookMutation.isPending}
                webhookDescription={webhook.description}
              />
            </>
          );
        };

        return <ActionCell />;
      },
    },
  ];
};
