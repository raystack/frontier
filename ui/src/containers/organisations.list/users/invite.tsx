import { zodResolver } from "@hookform/resolvers/zod";
import { Button, Flex, Sheet, Text } from "@raystack/apsara";
import { CSSProperties, useCallback } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router-dom";
import * as z from "zod";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";
import { Form } from "@radix-ui/react-form";
const inviteSchema = z.object({
  type: z.string(),
  team: z.string().optional(),
  emails: z
    .string()
    .transform((value) => value.split(",").map((str) => str.trim()))
    .pipe(z.string().email()),
});

type InviteSchemaType = z.infer<typeof inviteSchema>;

export default function InviteUsers() {
  const { organisationId } = useParams();
  const navigate = useNavigate();

  const methods = useForm<InviteSchemaType>({
    resolver: zodResolver(inviteSchema),
    defaultValues: {},
  });

  const onOpenChange = useCallback(() => {
    navigate(`/organisations/${organisationId}/users`);
  }, [organisationId]);

  const onSubmit = async (data: InviteSchemaType) => {
    console.log(data);
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
            <SheetHeader
              title="Invite users"
              onClick={onOpenChange}
            ></SheetHeader>
            <Flex direction="column" gap="large" style={styles.main}>
              <textarea
                style={styles.textarea}
                placeholder="Enter comma separated emails like abc@domain.com, bcd@domain.com"
              />
            </Flex>
            <SheetFooter>
              <Button type="submit" variant="primary" size={"medium"}>
                <Text
                  style={{
                    color: "var(--foreground-inverted)",
                  }}
                >
                  Invite users
                </Text>
              </Button>
            </SheetFooter>
          </Form>
        </FormProvider>
      </Sheet.Content>
    </Sheet>
  );
}

const styles: Record<string, CSSProperties> = {
  main: { padding: "32px" },
  textarea: {
    margin: 0,
    outline: "none",
    padding: "var(--pd-8)",
    height: "auto",
    width: "auto",

    backgroundColor: "var(--background-base)",
    border: "0.5px solid var(--border-base)",
    boxShadow: "var(--shadow-xs)",
    borderRadius: "var(--br-4)",
    color: "var(--foreground-base)",
    resize: "vertical",
  },
};
