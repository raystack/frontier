import { zodResolver } from "@hookform/resolvers/zod";
import useSWRMutation from "swr/mutation";

import { Form } from "@radix-ui/react-form";
import { Button, Flex, Sheet, Text } from "@raystack/apsara";
import * as z from "zod";

import { useCallback } from "react";
import { FormProvider, useForm, UseFormRegister } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { updateOrganisation } from "~/api";
import { CustomFieldName } from "~/components/CustomField";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";

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
  const { trigger } = useSWRMutation(
    "/v1beta1/organizations",
    updateOrganisation,
    {}
  );

  const methods = useForm<OrganizationForm>({
    resolver: zodResolver(OrganizationSchema),
    defaultValues: {},
  });

  const onOpenChange = useCallback(() => {
    navigate("/console/organisations");
  }, []);

  const onSubmit = async (data: any) => {
    console.log(JSON.stringify(data));
    await trigger(data);
    navigate("/console/organisations");
    navigate(0);
  };

  console.log(methods);

  return (
    <Sheet open={true}>
      <Sheet.Content
        side="right"
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
              <Button type="submit" variant="primary">
                <Text
                  style={{
                    color: "var(--foreground-inverted)",
                  }}
                >
                  Add organisation
                </Text>
              </Button>
            </SheetFooter>
          </Form>
        </FormProvider>
      </Sheet.Content>
    </Sheet>
  );
}

type CustomFieldNameProps = {
  name: string;
  register: UseFormRegister<OrganizationForm>;
};

const styles = {
  main: { padding: "32px", width: "80%" },
};
