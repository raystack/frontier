import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";

import {
  Form,
  FormControl,
  FormField,
  FormLabel,
  FormSubmit,
} from "@radix-ui/react-form";
import { Button, Flex, Sheet } from "@raystack/apsara/v1";
import * as z from "zod";

import { V1Beta1Organization } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import { useCallback, useEffect, useState } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { CustomFieldName } from "~/components/CustomField";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";

// TODO: Setting this to 1000 initially till APIs support filters and sorting.
const page_size = 1000;
const page_num = 1;

const GroupSchema = z.object({
  name: z
    .string()
    .trim()
    .min(2, { message: "Must be 2 or more characters long" }),
  slug: z
    .string()
    .trim()
    .toLowerCase()
    .min(2, { message: "Must be 2 or more characters long" }),
  orgId: z.string().trim(),
});
export type GroupForm = z.infer<typeof GroupSchema>;

export default function NewGroup() {
  const [organisation, setOrganisation] = useState();
  const [organisations, setOrganisations] = useState<V1Beta1Organization[]>([]);
  const navigate = useNavigate();
  const { client } = useFrontier();

  async function getOrganizations() {
    try {
      const res = await client?.adminServiceListAllOrganizations({
        page_num,
        page_size,
      });
      const organizations = res?.data.organizations ?? [];
      setOrganisations(organizations);
    } catch (error) {
      console.error(error);
    }
  }

  useEffect(() => {
    getOrganizations();
  }, []);

  const methods = useForm<GroupForm>({
    resolver: zodResolver(GroupSchema),
    defaultValues: {},
  });

  const onOpenChange = useCallback(() => {
    navigate("/groups");
  }, []);

  const onSubmit = async (data: any) => {
    if (!organisation) return;
    try {
      await client?.frontierServiceCreateGroup(organisation, data);
      toast.success("members added");
      navigate("/groups");
      navigate(0);
    } catch (error: any) {
      toast.error("Something went wrong", {
        description: error.message,
      });
    }
  };

  const onChange = (e: any) => {
    setOrganisation(e.target.value);
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
              title="Add new group"
              onClick={onOpenChange}
              data-test-id="admin-ui-add-new-group-header-btn"
            ></SheetHeader>
            <Flex direction="column" gap="large" style={styles.main}>
              <CustomFieldName
                name="name"
                register={methods.register}
                control={methods.control}
              />
              <CustomFieldName
                name="slug"
                register={methods.register}
                control={methods.control}
              />
              <FormField name="orgId" style={styles.formfield}>
                <Flex
                  style={{
                    marginBottom: "$1",
                    alignItems: "baseline",
                    justifyContent: "space-between",
                  }}
                >
                  <FormLabel
                    style={{
                      fontSize: "11px",
                      color: "#6F6F6F",
                      lineHeight: "16px",
                    }}
                  >
                    Organisation Id
                  </FormLabel>
                </Flex>
                <FormControl asChild>
                  <select
                    {...methods.register("orgId")}
                    style={styles.select}
                    onChange={onChange}
                    data-test-id="admin-ui-create-group-btn"
                  >
                    {organisations.map((org: V1Beta1Organization) => (
                      <option value={org.id} key={org.id}>
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
                  data-test-id="admin-ui-add-group-footer-btn"
                >
                  Add group
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
