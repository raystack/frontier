import { useContext } from "react";
import { useMatch, useParams, useNavigate } from "react-router-dom";
import { WebhooksView } from "@raystack/frontier/admin";
import { AppContext } from "~/contexts/App";

export function WebhooksPage() {
  const { config } = useContext(AppContext);
  const { webhookId } = useParams();
  const navigate = useNavigate();
  const isCreate = useMatch("/webhooks/create");

  const enableDelete = config?.webhooks?.enable_delete ?? false;

  return (
    <WebhooksView
      selectedWebhookId={webhookId}
      createOpen={!!isCreate}
      onCloseDetail={() => navigate("/webhooks")}
      onSelectWebhook={(id: string) => navigate(`/webhooks/${encodeURIComponent(id)}`)}
      onOpenCreate={() => navigate("/webhooks/create")}
      enableDelete={enableDelete}
    />
  );
}
