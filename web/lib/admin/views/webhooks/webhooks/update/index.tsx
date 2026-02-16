import { useCallback, useEffect } from "react";
import { Button, Flex, Sheet } from "@raystack/apsara";
import { useNavigate, useParams } from "react-router-dom";
import { SheetHeader } from "../../../../components/SheetHeader";
import { SheetFooter } from "../../../../components/SheetFooter";
import * as z from "zod";
import { FormProvider, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, FormSubmit } from "@radix-ui/react-form";
import { CustomFieldName } from "../../../../components/CustomField";
import events from "../../../../utils/webhook-events";
import { toast } from "sonner";
import { useMutation } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  type WebhookRequestBody,
  WebhookRequestBodySchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useWebhookQueries } from "../hooks/useWebhookQueries";

const UpdateWebhookSchema = z.object({
  url: z.string().trim().url(),
  description: z
    .string()
    .trim()
    .min(3, { message: "Must be 10 or more characters long" }),
  state: z.boolean().default(false),
  subscribed_events: z.array(z.string()).default([]),
});

export type UpdateWebhook = z.infer<typeof UpdateWebhookSchema>;

export type UpdateWebhooksProps = {
  webhookId?: string;
  onClose?: () => void;
};

export default function UpdateWebhooks({ webhookId: webhookIdProp, onClose: onCloseProp }: UpdateWebhooksProps = {}) {
  const navigate = useNavigate();
  const { webhookId: webhookIdParam } = useParams();
  const webhookId = webhookIdProp ?? webhookIdParam ?? "";

  const {
    listWebhooks: {
      data: webhooksResponse,
      isLoading: isWebhookLoading,
    },
    invalidateWebhooksList,
  } = useWebhookQueries();

  const onClose = useCallback(() => {
    if (onCloseProp) onCloseProp();
    else navigate("/webhooks");
  }, [navigate, onCloseProp]);

  const methods = useForm<UpdateWebhook>({
    resolver: zodResolver(UpdateWebhookSchema),
    defaultValues: {},
  });

  const webhook = webhooksResponse?.webhooks?.find(
    (wb) => wb?.id === webhookId,
  );

  const { mutateAsync: updateWebhook, isPending: isSubmitting } = useMutation(
    AdminServiceQueries.updateWebhook,
  );

  const onSubmit = async (data: UpdateWebhook) => {
    try {
      const body: WebhookRequestBody = create(WebhookRequestBodySchema, {
        url: data.url,
        description: data.description,
        state: data.state ? "enabled" : "disabled",
        subscribedEvents: data.subscribed_events || [],
        headers: {},
      });

      const resp = await updateWebhook({
        id: webhookId,
        body,
      });

      if (resp?.webhook) {
        toast.success("Webhook updated");
        await invalidateWebhooksList();
        onClose();
      }
    } catch (err) {
      console.error("Failed to update webhook:", err);
      toast.error("Something went wrong");
    }
  };

  useEffect(() => {
    if (webhook) {
      methods.reset({
        url: webhook.url,
        description: webhook.description,
        subscribed_events: webhook.subscribedEvents || [],
        state: webhook.state === "enabled",
      });
    }
  }, [webhook, methods.reset]);

  return (
    <Sheet open={true}>
      <Sheet.Content
        side="right"
        // @ts-ignore
        style={{
          width: "30vw",
          padding: 0,
          boxShadow: "var(--rs-shadow-soft)",
        }}
        close={false}
      >
        <FormProvider {...methods}>
          <Form onSubmit={methods.handleSubmit(onSubmit)}>
            <SheetHeader
              title="Update Webhook"
              onClick={onClose}
              data-testid="admin-update-webhook-close-btn"
            />
            <Flex
              direction="column"
              gap={9}
              style={styles.main}
              key={webhook?.id}
            >
              <CustomFieldName
                name="url"
                defaultValue={webhook?.url}
                register={methods.register}
                control={methods.control}
                variant="textarea"
                style={{ width: "100%" }}
                isLoading={isWebhookLoading}
              />
              <CustomFieldName
                name="description"
                defaultValue={webhook?.description}
                register={methods.register}
                control={methods.control}
                variant="textarea"
                style={{ width: "100%" }}
                isLoading={isWebhookLoading}
              />
              <CustomFieldName
                name="subscribed_events"
                defaultValue={webhook?.subscribedEvents}
                register={methods.register}
                control={methods.control}
                variant="multiselect"
                options={events.map((e: string) => ({ label: e, value: e }))}
                isLoading={isWebhookLoading}
              />
              <CustomFieldName
                name="state"
                defaultChecked={webhook?.state === "enabled"}
                register={methods.register}
                control={methods.control}
                variant="switch"
                isLoading={isWebhookLoading}
              />
            </Flex>
            <SheetFooter>
              <FormSubmit asChild>
                <Button
                  style={{ height: "inherit" }}
                  disabled={isSubmitting || isWebhookLoading}
                  data-test-id="admin-submit-btn"
                  loading={isSubmitting}
                  loaderText="Updating..."
                >
                  Update Webhook
                </Button>
              </FormSubmit>
            </SheetFooter>
          </Form>
        </FormProvider>
      </Sheet.Content>
    </Sheet>
  );
}

const styles = {
  main: {
    padding: "32px",
    margin: 0,
    height: "calc(100vh - 125px)",
    overflow: "auto",
  },
  formfield: {
    width: "80%",
    marginBottom: "40px",
  },
  select: {
    height: "32px",
    borderRadius: "8px",
    padding: "8px",
    border: "none",
    backgroundColor: "transparent",
    "&:active,&:focus": {
      border: "none",
      outline: "none",
      boxShadow: "none",
    },
  },
};
