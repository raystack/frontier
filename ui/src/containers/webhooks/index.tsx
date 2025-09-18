import { Flex, DataTable, EmptyState } from "@raystack/apsara";
import { getColumns } from "./columns";
import { WebhooksHeader } from "./header";
import { Outlet, useNavigate } from "react-router-dom";
import styles from "./webhooks.module.css";
import { useQuery } from "@connectrpc/connect-query";
import { AdminServiceQueries } from "@raystack/proton/frontier";
import { ExclamationTriangleIcon } from "@radix-ui/react-icons";

export default function WebhooksList() {
  const navigate = useNavigate();

  const {
    data: webhooksResponse,
    isLoading,
    error,
    isError,
  } = useQuery(
    AdminServiceQueries.listWebhooks,
    {},
    {
      staleTime: 0,
      refetchOnWindowFocus: false,
    },
  );

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

  const columns = getColumns({ openEditPage });
  return (
    <DataTable
      data={webhooks}
      columns={columns}
      isLoading={isLoading}
      defaultSort={{ name: "created_at", order: "desc" }}
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
