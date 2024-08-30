import { zodResolver } from "@hookform/resolvers/zod";
import { Form, FormSubmit } from "@radix-ui/react-form";
import { Button, Flex, Sheet, Text } from "@raystack/apsara";
import * as z from "zod";

import { useFrontier } from "@raystack/frontier/react";
import { useCallback } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { CustomFieldName } from "~/components/CustomField";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";

const UserSchema = z.object({
  title: z
    .string()
    .trim()
    .min(3, { message: "Must be 3 or more characters long" }),
  email: z.string().email(),
});
export type UserForm = z.infer<typeof UserSchema>;

export default function NewUser() {
  const { client } = useFrontier();
  const navigate = useNavigate();

  const methods = useForm<UserForm>({
    resolver: zodResolver(UserSchema),
    defaultValues: {},
  });

  const onOpenChange = useCallback(() => {
    navigate("/users");
  }, []);

  const onSubmit = async (data: any) => {
    try {
      await client?.frontierServiceCreateUser(data);
      toast.success("user added");
      navigate("/users");
      navigate(0);
    } catch (error: any) {
      toast.error("Something went wrong", {
        description: error.message,
      });
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
            <SheetHeader
              title="Add new user"
              onClick={onOpenChange}
              data-test-id="admin-ui-sheet-header"
            ></SheetHeader>
            <Flex direction="column" gap="large" style={styles.main}>
              <CustomFieldName
                name="title"
                register={methods.register}
                control={methods.control}
              />
              <CustomFieldName
                name="email"
                register={methods.register}
                control={methods.control}
              />
            </Flex>
            <SheetFooter>
              <FormSubmit asChild>
                <Button variant="primary" style={{ height: "inherit" }} data-test-id="admin-ui-add-user-btn">
                  <Text
                    style={{
                      fontSize: "14px",
                      color: "white",
                      fontWeight: "normal",
                      lineHeight: "20px",
                    }}
                  >
                    Add user
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
  main: { padding: "32px", width: "80%", margin: 0 },
  formfield: {
    marginBottom: "40px",
  },
};
