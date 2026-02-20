import { Flex, DataTable, EmptyState } from "@raystack/apsara";
import { useCallback } from "react";
import { getColumns } from "./columns";
import { WebhooksHeader } from "./header";
import styles from "./webhooks.module.css";
import { useWebhookQueries } from "./hooks/useWebhookQueries";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import CreateWebhooks from "./create";
import UpdateWebhooks from "./update";

export type WebhooksViewProps = {
  selectedWebhookId?: string;
  createOpen?: boolean;
  onCloseDetail?: () => void;
  onSelectWebhook?: (id: string) => void;
  onOpenCreate?: () => void;
  enableDelete?: boolean;
};

export default function WebhooksView({
  selectedWebhookId,
  createOpen,
  onCloseDetail,
  onSelectWebhook,
  onOpenCreate,
  enableDelete = false,
}: WebhooksViewProps = {}) {
  const {
    listWebhooks: {
      data: webhooksResponse,
      isLoading,
      error,
      isError,
    },
    deleteWebhookMutation,
  } = useWebhookQueries();

  const webhooks = webhooksResponse?.webhooks || [];

  const openEditPage = useCallback(
    (id: string) => (onSelectWebhook ? onSelectWebhook(id) : undefined),
    [onSelectWebhook]
  );

  if (isError) {
    console.error("ConnectRPC Error:", error);
    return (
      <EmptyState
        icon={<ExclamationTriangleIcon />}
        heading="Error Loading Webhooks"
        subHeading={
          error?.message ||
          "Something went wrong while loading webhooks. Please try again."
        }
      />
    );
  }

  const columns = getColumns({
    openEditPage: (id) => {
      openEditPage(id);
    },
    deleteWebhookMutation,
    enableDelete,
  });

  return (
    <>
      <DataTable
        data={webhooks}
        columns={columns}
        isLoading={isLoading}
        defaultSort={{ name: "createdAt", order: "desc" }}
        mode="client"
      >
        <Flex direction="column" className={styles.tableWrapper}>
          <WebhooksHeader onOpenCreate={onOpenCreate} />
          <DataTable.Content
            classNames={{ root: styles.tableRoot, table: styles.table }}
          />
        </Flex>
      </DataTable>
      {createOpen && <CreateWebhooks onClose={onCloseDetail} />}
      {selectedWebhookId && (
        <UpdateWebhooks
          webhookId={selectedWebhookId}
          onClose={onCloseDetail}
        />
      )}
    </>
  );
}
