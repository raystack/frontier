import { Flex, DataTable, EmptyState } from "@raystack/apsara";
import { getColumns } from "./columns";
import { WebhooksHeader } from "./header";
import { Outlet, useNavigate } from "react-router-dom";
import styles from "./webhooks.module.css";
import { useWebhookQueries } from "./hooks/useWebhookQueries";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

export default function WebhooksList() {
  const navigate = useNavigate();
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

  function openEditPage(id: string) {
    navigate(`/webhooks/${id}`);
  }

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

  const columns = getColumns({ openEditPage, deleteWebhookMutation });
  return (
    <DataTable
      data={webhooks}
      columns={columns}
      isLoading={isLoading}
      defaultSort={{ name: "createdAt", order: "desc" }}
      mode="client"
    >
      <Flex direction="column" className={styles.tableWrapper}>
        <WebhooksHeader />
        <DataTable.Content
          classNames={{ root: styles.tableRoot, table: styles.table }}
        />
        <Outlet />
      </Flex>
    </DataTable>
  );
}
