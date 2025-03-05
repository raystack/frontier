import { useCallback, useEffect, useState } from "react";
import { Button, Flex, Sheet, Text } from "@raystack/apsara/v1";
import { useNavigate, useParams } from "react-router-dom";
import { SheetHeader } from "~/components/sheet/header";
import * as z from "zod";
import { FormProvider, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, FormSubmit } from "@radix-ui/react-form";
import { CustomFieldName } from "~/components/CustomField";
import events from "~/utils/webhook_events";
import { SheetFooter } from "~/components/sheet/footer";
import { useFrontier } from "@raystack/frontier/react";
import { V1Beta1Webhook, V1Beta1WebhookRequestBody } from "@raystack/frontier";
import { toast } from "sonner";

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

export default function UpdateWebhooks() {
  const navigate = useNavigate();
  const { client } = useFrontier();

  const [isSubmitting, setIsSubmitting] = useState(false);

  const [isWebhookLoading, setIsWebhookLoading] = useState(false);
  const [webhook, setWebhook] = useState<V1Beta1Webhook>();

  const { webhookId = "" } = useParams();

  const onClose = useCallback(() => {
    navigate("/webhooks");
  }, [navigate]);

  const methods = useForm<UpdateWebhook>({
    resolver: zodResolver(UpdateWebhookSchema),
    defaultValues: {},
  });

  const onSubmit = async (data: UpdateWebhook) => {
    try {
      setIsSubmitting(true);
      const body: V1Beta1WebhookRequestBody = {
        ...data,
        state: data.state ? "enabled" : "disabled",
      };
      const resp = await client?.adminServiceUpdateWebhook(webhookId, {
        body,
      });

      if (resp?.data?.webhook) {
        toast.success("Webhook updated");
      }
    } catch (err) {
      toast.error("Something went wrong");
    } finally {
      setIsSubmitting(false);
    }
  };

  useEffect(() => {
    async function getWebhookDetails(id: string) {
      try {
        setIsWebhookLoading(true);
        const resp = await client?.adminServiceListWebhooks();
        const webhooks = resp?.data?.webhooks || [];
        const webhookData = webhooks?.find((wb) => wb?.id === id);
        setWebhook(webhookData);
        methods?.reset({
          ...webhookData,
          state: webhookData?.state === "enabled",
        });
      } catch (err) {
        console.error(err);
      } finally {
        setIsWebhookLoading(false);
      }
    }

    if (webhookId) {
      getWebhookDetails(webhookId);
    }
  }, [client, methods, webhookId]);

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
        <FormProvider {...methods}>
          <Form onSubmit={methods.handleSubmit(onSubmit)}>
            <SheetHeader
              title="Update Webhook"
              onClick={onClose}
              data-test-id="admin-ui-update-webhook-close-btn"
            />
            <Flex direction="column" gap="large" style={styles.main}>
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
                defaultValue={webhook?.subscribed_events}
                register={methods.register}
                control={methods.control}
                variant="multiselect"
                options={events.map((e) => ({ label: e, value: e }))}
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
                  data-test-id="admin-ui-submit-btn"
                >
                  <Text size={4} variant={"emphasis"}>
                    {isSubmitting ? "Updating..." : "Update Webhook"}
                  </Text>
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
