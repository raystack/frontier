import { Flex, DataTable } from "@raystack/apsara/v1";
import type { V1Beta1Webhook } from "@raystack/frontier";
import { useEffect, useState } from "react";
import { getColumns } from "./columns";
import { WebhooksHeader } from "./header";
import { Outlet, useNavigate } from "react-router-dom";
import { api } from "~/api";
import styles from "./webhooks.module.css";

export default function WebhooksList() {
  const navigate = useNavigate();
  const [webhooks, setWebhooks] = useState<V1Beta1Webhook[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    async function fetchWebhooks() {
      try {
        setIsLoading(true);
        const resp = await api?.adminServiceListWebhooks();
        const data = resp?.data?.webhooks || [];
        setWebhooks(data);
      } catch (err) {
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    }
    fetchWebhooks();
  }, []);

  function openEditPage(id: string) {
    navigate(`/webhooks/${id}`);
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
