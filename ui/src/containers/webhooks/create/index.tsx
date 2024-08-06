import { Sheet } from "@raystack/apsara";
import { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { SheetHeader } from "~/components/sheet/header";

export default function CreateWebhooks() {
  const navigate = useNavigate();

  const onOpenChange = useCallback(() => {
    navigate("/webhooks");
  }, [navigate]);

  return (
    <Sheet open={true}>
      <Sheet.Content
        side="right"
        // @ts-ignore
        style={{
          width: "30vw",
          padding: 0,
          borderRadius: "var(--pd-8)",
          boxShadow: "var(--shadow-sm)",
        }}
        close={false}
      >
        <SheetHeader
          title="Add new Webhook"
          onClick={onOpenChange}
        ></SheetHeader>
      </Sheet.Content>
    </Sheet>
  );
}
