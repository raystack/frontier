import { Button, Flex, Sheet, Text, TextField } from "@raystack/apsara";
import { useCallback, useState } from "react";
import { useNavigate } from "react-router-dom";
import { SheetHeader } from "~/components/sheet/header";
import * as z from "zod";
import { Controller, FormProvider, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, FormSubmit } from "@radix-ui/react-form";
import { CustomFieldName } from "~/components/CustomField";
import events from "./events";
import { SheetFooter } from "~/components/sheet/footer";
import { useFrontier } from "@raystack/frontier/react";
import { V1Beta1WebhookRequestBody } from "@raystack/frontier";
import { toast } from "sonner";

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
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { client } = useFrontier();

  const onOpenChange = useCallback(() => {
    navigate("/webhooks");
  }, [navigate]);

  const methods = useForm<NewWebhook>({
    resolver: zodResolver(NewWebookSchema),
    defaultValues: {},
  });

  const onSubmit = async (data: NewWebhook) => {
    try {
      setIsSubmitting(true);
      const body: V1Beta1WebhookRequestBody = {
        ...data,
        state: data.state ? "enabled" : "disabled",
      };
      const resp = await client?.adminServiceCreateWebhook({
        body,
      });

      if (resp?.data?.webhook) {
        toast.success("Webhook created");
        onOpenChange();
      }
    } catch (err) {
      toast.error("Something went wrong");
    } finally {
      setIsSubmitting(false);
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
          borderRadius: "var(--pd-8)",
          boxShadow: "var(--shadow-sm)",
        }}
        close={false}
      >
        <FormProvider {...methods}>
          <Form onSubmit={methods.handleSubmit(onSubmit)}>
            <SheetHeader title="Add new Webhook" onClick={onOpenChange} />
            <Flex direction="column" gap="large" style={styles.main}>
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
                  variant="primary"
                  style={{ height: "inherit" }}
                  disabled={isSubmitting}
                >
                  <Text
                    size={4}
                    style={{ color: "var(--foreground-inverted)" }}
                  >
                    {isSubmitting ? "Adding..." : "Add Webhook"}
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
