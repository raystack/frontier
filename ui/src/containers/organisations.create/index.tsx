import { zodResolver } from "@hookform/resolvers/zod";

import { Form } from "@radix-ui/react-form";
import { Button, Flex, Sheet } from "@raystack/apsara/v1";
import * as z from "zod";

import { useCallback } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { CustomFieldName } from "~/components/CustomField";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";
import { api } from "~/api";

const OrganizationSchema = z.object({
  title: z
    .string()
    .trim()
    .min(3, { message: "Must be 3 or more characters long" }),
  name: z
    .string()
    .trim()
    .toLowerCase()
    .min(3, { message: "Must be 3 or more characters long" }),
});
export type OrganizationForm = z.infer<typeof OrganizationSchema>;

export default function NewOrganisation() {
  const navigate = useNavigate();

  const methods = useForm<OrganizationForm>({
    resolver: zodResolver(OrganizationSchema),
    defaultValues: {},
  });

  const onOpenChange = useCallback(() => {
    navigate("/organisations");
  }, []);

  const onSubmit = async (data: any) => {
    try {
      await api?.frontierServiceCreateOrganization(data);
      toast.success("organisation added");
      navigate("/organisations");
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
              title="Add new organisation"
              onClick={onOpenChange}
              data-test-id="admin-ui-add-new-organisation-btn"
            ></SheetHeader>
            <Flex direction="column" gap="large" style={styles.main}>
              <CustomFieldName
                name="title"
                register={methods.register}
                control={methods.control}
              />
              <CustomFieldName
                name="name"
                register={methods.register}
                control={methods.control}
              />
            </Flex>
            <SheetFooter>
              <Button
                type="submit"
                data-test-id="admin-ui-create-org-footer-btn"
              >
                Add organisation
              </Button>
            </SheetFooter>
          </Form>
        </FormProvider>
      </Sheet.Content>
    </Sheet>
  );
}

const styles = {
  main: { padding: "32px", width: "80%" },
};
