import { DataTable } from "@raystack/apsara";
import { Flex } from "@raystack/apsara/v1";
import { V1Beta1Webhook } from "@raystack/frontier";
import { useEffect, useState } from "react";
import { getColumns } from "./columns";
import { WebhooksHeader } from "./header";
import { Outlet, useNavigate } from "react-router-dom";
import { api } from "~/api";

export default function WebhooksList() {
  const tableStyle = { width: "100%" };
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
    <Flex direction="row" style={{ height: "100%", width: "100%" }}>
      <DataTable
        data={webhooks}
        columns={columns}
        isLoading={isLoading}
        parentStyle={{ height: "calc(100vh - 60px)" }}
        style={tableStyle}
      >
        <DataTable.Toolbar>
          <WebhooksHeader />
          <DataTable.FilterChips style={{ padding: "8px 24px" }} />
        </DataTable.Toolbar>
        <DataTable.DetailContainer>
          <Outlet />
        </DataTable.DetailContainer>
      </DataTable>
    </Flex>
  );
}
