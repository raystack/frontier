import { DataTable, Flex } from "@raystack/apsara";
import { V1Beta1Webhook } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useEffect, useState } from "react";
import { getColumns } from "./columns";
import { date } from "zod";

export function WebhooksList() {
  const tableStyle = { width: "100%" };
  const { client } = useFrontier();
  const [webhooks, setWebhooks] = useState<V1Beta1Webhook[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    async function fetchWebhooks() {
      try {
        setIsLoading(true);
        const resp = await client?.adminServiceListWebhooks();
        const data = resp?.data?.webhooks || [];
        setWebhooks(data);
      } catch (err) {
        console.error(err);
      } finally {
        setIsLoading(false);
      }
    }
    fetchWebhooks();
  }, [client]);

  const columns = getColumns();
  return (
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={webhooks}
        columns={columns}
        isLoading={isLoading}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
      </DataTable>
    </Flex>
  );
}
