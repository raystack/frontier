import { useCallback } from "react";
import { Button, Flex, Drawer, toastManager } from "@raystack/apsara-v1";
import { SheetHeader } from "../../../../components/SheetHeader";
import { SheetFooter } from "../../../../components/SheetFooter";
import * as z from "zod";
import { FormProvider, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, FormSubmit } from "@radix-ui/react-form";
import { CustomFieldName } from "../../../../components/CustomField";
import events from "../../../../utils/webhook-events";
import { useMutation } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  type WebhookRequestBody,
  WebhookRequestBodySchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
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

export type CreateWebhooksProps = {
  open?: boolean;
  onClose?: () => void;
};

export default function CreateWebhooks({ open = false, onClose: onCloseProp }: CreateWebhooksProps = {}) {
  const { invalidateWebhooksList } = useWebhookQueries();

  const onOpenChange = useCallback(() => {
    onCloseProp?.();
  }, [onCloseProp]);

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
        toastManager.add({ title: "Webhook created", type: "success" });
        await invalidateWebhooksList();
        onOpenChange();
      }
    } catch (err) {
      console.error("Failed to create webhook:", err);
      toastManager.add({ title: "Something went wrong", type: "error" });
    }
  };

  return (
    <Drawer open={open}>
      <Drawer.Content
        side="right"
        // @ts-ignore
        style={{
          width: "30vw",
          padding: 0,
          boxShadow: "var(--rs-shadow-soft)",
        }}
        showCloseButton={false}
      >
        <FormProvider {...methods}>
          <Form onSubmit={methods.handleSubmit(onSubmit)}>
            <SheetHeader
              title="Add new Webhook"
              onClick={onOpenChange}
              data-test-id="admin-add-new-webhook-btn"
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
                options={events.map((e: string) => ({ label: e, value: e }))}
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
                  data-test-id="admin-submit-btn"
                  loading={isSubmitting}
                  loaderText="Adding..."
                >
                  Add Webhook
                </Button>
              </FormSubmit>
            </SheetFooter>
          </Form>
        </FormProvider>
      </Drawer.Content>
    </Drawer>
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
