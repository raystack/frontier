import { zodResolver } from "@hookform/resolvers/zod";

import { Form } from "@radix-ui/react-form";
import { Button, Flex, Sheet } from "@raystack/apsara/v1";
import * as z from "zod";

import { useFrontier } from "@raystack/frontier/react";
import { useCallback } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router-dom";
import { toast } from "sonner";
import { CustomFieldName } from "~/components/CustomField";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";

const ServiceUserSchema = z.object({
  title: z
    .string()
    .trim()
    .min(3, { message: "Must be 3 or more characters long" }),
});
export type ServiceUserForm = z.infer<typeof ServiceUserSchema>;

export default function NewServiceUsers() {
  const { client } = useFrontier();
  const navigate = useNavigate();
  let { organisationId } = useParams();

  const methods = useForm<ServiceUserForm>({
    resolver: zodResolver(ServiceUserSchema),
    defaultValues: {},
  });

  const onOpenChange = useCallback(() => {
    navigate(`/organisations/${organisationId}/serviceusers`);
  }, [navigate, organisationId]);

  const onSubmit = async (data: ServiceUserForm) => {
    try {
      if (organisationId) {
        await client?.frontierServiceCreateServiceUser(organisationId, {
          body: data,
        });
        toast.success("service user added");
        onOpenChange();
        navigate(0);
      }
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
              title="Add new service user"
              onClick={onOpenChange}
              data-test-id="admin-ui-add-new-service-user-btn"
            ></SheetHeader>
            <Flex direction="column" gap="large" style={styles.main}>
              <CustomFieldName
                name="title"
                register={methods.register}
                control={methods.control}
              />
            </Flex>
            <SheetFooter>
              <Button
                type="submit"
                data-test-id="admin-ui-add-new-service-user-footer-btn"
              >
                Add new service user
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
