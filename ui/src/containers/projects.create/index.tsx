import { zodResolver } from "@hookform/resolvers/zod";

import {
  Form,
  FormControl,
  FormField,
  FormLabel,
  FormSubmit,
} from "@radix-ui/react-form";
import { Button, Flex, Sheet, Text } from "@raystack/apsara/v1";
import * as z from "zod";

import { V1Beta1Organization } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useCallback, useEffect, useState } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { CustomFieldName } from "~/components/CustomField";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";

// TODO: Setting this to 1000 initially till APIs support filters and sorting.
const page_size = 1000;
const page_num = 1;

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
  const { client } = useFrontier();
  const [organisations, setOrganisations] = useState<V1Beta1Organization[]>([]);

  async function getOrganizations() {
    try {
      const res = await client?.adminServiceListAllOrganizations({
        page_num,
        page_size,
      });
      const organizations = res?.data?.organizations ?? [];
      setOrganisations(organizations);
    } catch (error) {
      console.error(error);
    }
  }

  useEffect(() => {
    getOrganizations();
  }, []);

  const methods = useForm<ProjectForm>({
    resolver: zodResolver(ProjectSchema),
    defaultValues: {},
  });

  const onOpenChange = useCallback(() => {
    navigate("/projects");
  }, []);

  const onSubmit = async (data: any) => {
    try {
      await client?.frontierServiceCreateProject(data);
      toast.success("project added");
      navigate("/projects");
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
              title="Add new project"
              onClick={onOpenChange}
              data-test-id="admin-ui-add-new-project-header"
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
                    {organisations.map((org: V1Beta1Organization) => (
                      <option key={org.id} value={org.id}>
                        {org.name}
                      </option>
                    ))}
                  </select>
                </FormControl>
              </FormField>
            </Flex>
            <SheetFooter>
              <FormSubmit asChild>
                <Button
                  style={{ height: "inherit" }}
                  data-test-id="admin-ui-add-project-btn"
                >
                  Add project
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
