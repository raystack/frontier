import { Button, Flex, DataTable, EmptyState } from "@raystack/apsara-v1";
import { useCallback, type ReactNode } from "react";
import { getColumns } from "./columns";
import styles from "./webhooks.module.css";
import { useWebhookQueries } from "./hooks/useWebhookQueries";
import { ExclamationTriangleIcon, PlusIcon } from "@radix-ui/react-icons";
import { PageHeader } from "../../../components/PageHeader";
import CreateWebhooks from "./create";
import UpdateWebhooks from "./update";

export type WebhooksViewProps = {
  /** When set, opens the update panel for this webhook. */
  selectedWebhookId?: string;
  /** When true, opens the create webhook panel. */
  createOpen?: boolean;
  /** Called when the detail/create panel is closed. */
  onCloseDetail?: () => void;
  /** Called when a user clicks a webhook row. Use to update the URL or local state. */
  onSelectWebhook?: (id: string) => void;
  /** Called when the "Create" button is clicked. Use to update the URL or open the create panel. */
  onOpenCreate?: () => void;
  /** When true, shows the delete option for webhooks. Defaults to `false`. */
  enableDelete?: boolean;
  /** Icon rendered in the page header next to the title. */
  icon?: ReactNode;
};

export default function WebhooksView({
  selectedWebhookId,
  createOpen,
  onCloseDetail,
  onSelectWebhook,
  onOpenCreate,
  enableDelete = false,
  icon,
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
          <PageHeader
            title="Webhooks"
            icon={icon}
            breadcrumb={[]}
            className={styles.header}
          >
            <DataTable.Search placeholder="Search webhooks..." size="small" />
            <Button
              size="small"
              variant="text"
              color="neutral"
              leadingIcon={<PlusIcon />}
              data-test-id="admin-create-webhook-btn"
              onClick={() => onOpenCreate?.()}
            >
              New Webhook
            </Button>
          </PageHeader>
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
