import { zodResolver } from "@hookform/resolvers/zod";
import useSWR from "swr";
import useSWRMutation from "swr/mutation";

import {
  Form,
  FormControl,
  FormField,
  FormLabel,
  FormSubmit,
} from "@radix-ui/react-form";
import { Button, Flex, Sheet, Text } from "@raystack/apsara";
import * as z from "zod";

import { useCallback } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { update } from "~/api";
import { CustomFieldName } from "~/components/CustomField";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";
import { Organisation } from "~/types/organisation";
import { fetcher } from "~/utils/helper";

const ProjectSchema = z.object({
  title: z
    .string()
    .trim()
    .min(3, { message: "Must be 3 or more characters long" }),
  name: z
    .string()
    .trim()
    .toLowerCase()
    .min(3, { message: "Must be 3 or more characters long" }),
  orgId: z.string().trim(),
});
export type ProjectForm = z.infer<typeof ProjectSchema>;

export default function NewProject() {
  const navigate = useNavigate();
  const { data, error } = useSWR("/v1beta1/admin/organizations", fetcher);
  const { trigger } = useSWRMutation("/v1beta1/projects", update, {});
  const { organizations = [] } = data || { organizations: [] };

  const methods = useForm<ProjectForm>({
    resolver: zodResolver(ProjectSchema),
    defaultValues: {},
  });

  const onOpenChange = useCallback(() => {
    navigate("/console/projects");
  }, []);

  const onSubmit = async (data: any) => {
    await trigger(data);
    navigate("/console/projects");
    navigate(0);
  };

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
              title="Add new project"
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
              <FormField name="orgId" style={styles.formfield}>
                <Flex
                  style={{
                    alignItems: "baseline",
                    justifyContent: "space-between",
                  }}
                >
                  <FormLabel>Organisation Id</FormLabel>
                </Flex>
                <FormControl asChild>
                  <select {...methods.register("orgId")}>
                    {organizations.map((org: Organisation) => (
                      <option value={org.id}>{org.name}</option>
                    ))}
                  </select>
                </FormControl>
              </FormField>
            </Flex>
            <SheetFooter>
              <FormSubmit asChild>
                <Button variant="primary" style={{ height: "inherit" }}>
                  <Text
                    size={4}
                    style={{ color: "var(--foreground-inverted)" }}
                  >
                    Add project
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
