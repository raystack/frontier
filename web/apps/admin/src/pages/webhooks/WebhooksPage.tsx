import { useContext } from "react";
import { useMatch, useParams, useNavigate } from "react-router-dom";
import { WebhooksView } from "@raystack/frontier/admin";
import { AppContext } from "~/contexts/App";
import WebhooksIcon from "~/assets/icons/webhooks.svg?react";

export default function WebhooksPage() {
  const { config } = useContext(AppContext);
  const { webhookId } = useParams();
  const navigate = useNavigate();
  const isCreate = useMatch("/webhooks/create");

  // View-only unless `webhooks.enable_actions` is enabled in the deployment config.
  const enableActions = config?.webhooks?.enable_actions ?? false;

  return (
    <WebhooksView
      selectedWebhookId={webhookId}
      createOpen={!!isCreate}
      onCloseDetail={() => navigate("/webhooks")}
      onSelectWebhook={(id: string) => navigate(`/webhooks/${encodeURIComponent(id)}`)}
      onOpenCreate={() => navigate("/webhooks/create")}
      enableActions={enableActions}
      icon={<WebhooksIcon />}
    />
  );
}
