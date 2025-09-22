import { useCallback } from "react";
import { Button, Flex, Sheet } from "@raystack/apsara";
import { useNavigate } from "react-router-dom";
import { SheetHeader } from "~/components/sheet/header";
import * as z from "zod";
import { FormProvider, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, FormSubmit } from "@radix-ui/react-form";
import { CustomFieldName } from "~/components/CustomField";
import events from "~/utils/webhook_events";
import { SheetFooter } from "~/components/sheet/footer";
import { toast } from "sonner";
import { useMutation } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  type WebhookRequestBody,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { WebhookRequestBodySchema } from "@raystack/proton/frontier";
import { useWebhookQueries } from "../hooks/useWebhookQueries";

const NewWebookSchema = z.object({
  url: z.string().trim().url(),
  description: z
    .string()
    .trim()
    .min(3, { message: "Must be 10 or more characters long" }),
  state: z.boolean().default(false),
  subscribed_events: z.array(z.string()).default([]),
});

export type NewWebhook = z.infer<typeof NewWebookSchema>;

export default function CreateWebhooks() {
  const navigate = useNavigate();
  const { invalidateWebhooksList } = useWebhookQueries();

  const onOpenChange = useCallback(() => {
    navigate("/webhooks");
  }, [navigate]);

  const { mutateAsync: createWebhook, isPending: isSubmitting } = useMutation(
    AdminServiceQueries.createWebhook,
  );

  const methods = useForm<NewWebhook>({
    resolver: zodResolver(NewWebookSchema),
    defaultValues: {},
  });

  const onSubmit = async (data: NewWebhook) => {
    try {
      const body: WebhookRequestBody = create(WebhookRequestBodySchema, {
        url: data.url,
        description: data.description,
        state: data.state ? "enabled" : "disabled",
        subscribedEvents: data.subscribed_events || [],
        headers: {},
      });

      const resp = await createWebhook({ body });

      if (resp?.webhook) {
        toast.success("Webhook created");
        await invalidateWebhooksList();
        onOpenChange();
      }
    } catch (err) {
      console.error("Failed to create webhook:", err);
      toast.error("Something went wrong");
    }
  };

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
              title="Add new Webhook"
              onClick={onOpenChange}
              data-test-id="admin-ui-add-new-webhook-btn"
            />
            <Flex direction="column" gap={9} style={styles.main}>
              <CustomFieldName
                name="url"
                register={methods.register}
                control={methods.control}
                variant="textarea"
                style={{ width: "100%" }}
              />
              <CustomFieldName
                name="description"
                register={methods.register}
                control={methods.control}
                variant="textarea"
                style={{ width: "100%" }}
              />
              <CustomFieldName
                name="subscribed_events"
                register={methods.register}
                control={methods.control}
                variant="multiselect"
                options={events.map((e) => ({ label: e, value: e }))}
              />
              <CustomFieldName
                name="state"
                register={methods.register}
                control={methods.control}
                variant="switch"
              />
            </Flex>
            <SheetFooter>
              <FormSubmit asChild>
                <Button
                  style={{ height: "inherit" }}
                  disabled={isSubmitting}
                  data-test-id="admin-ui-submit-btn"
                  loading={isSubmitting}
                  loaderText="Adding..."
                >
                  Add Webhook
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
