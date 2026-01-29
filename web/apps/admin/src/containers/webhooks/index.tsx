import { Flex, DataTable, EmptyState } from "@raystack/apsara";
import { getColumns } from "./columns";
import { WebhooksHeader } from "./header";
import { Outlet, useNavigate } from "react-router-dom";
import styles from "./webhooks.module.css";
import { useWebhookQueries } from "./hooks/useWebhookQueries";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";
import { AppContext } from "~/contexts/App";
import { useContext } from "react";

export default function WebhooksList() {
  const navigate = useNavigate();
  const { config } = useContext(AppContext);
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

  const enableDelete = config.webhooks?.enable_delete ?? false;
  const columns = getColumns({ openEditPage, deleteWebhookMutation, enableDelete });
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
