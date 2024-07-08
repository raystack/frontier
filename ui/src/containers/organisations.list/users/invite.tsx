import { zodResolver } from "@hookform/resolvers/zod";
import { Button, Flex, Sheet, Text } from "@raystack/apsara";
import { CSSProperties, useCallback, useEffect, useState } from "react";
import { FormProvider, useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router-dom";
import * as z from "zod";
import { SheetFooter } from "~/components/sheet/footer";
import { SheetHeader } from "~/components/sheet/header";
import { Form } from "@radix-ui/react-form";
import { CustomFieldName } from "~/components/CustomField";
import { useFrontier } from "@raystack/frontier/react";
import { PERMISSIONS } from "~/utils/constants";
import Skeleton from "react-loading-skeleton";
import { V1Beta1Group, V1Beta1Role } from "@raystack/frontier";
import { toast } from "sonner";
import { HttpResponse } from "~/types/HttpResponse";

const inviteSchema = z.object({
  type: z
    .string()
    .transform((value) => value?.split(",").map((str) => str.trim())),
  team: z
    .string()
    .optional()
    .transform((value) => value?.split(",").map((str) => str.trim())),
  emails: z
    .string()
    .transform((value) => value.split(",").map((str) => str.trim()))
    .pipe(z.array(z.string().email())),
});

type InviteSchemaType = z.infer<typeof inviteSchema>;

export default function InviteUsers() {
  const { organisationId } = useParams();
  const { client } = useFrontier();
  const navigate = useNavigate();
  const [isDataLoading, setIsDataLoading] = useState(false);
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [groups, setGroups] = useState<V1Beta1Group[]>([]);

  const methods = useForm<InviteSchemaType>({
    resolver: zodResolver(inviteSchema),
    defaultValues: {},
  });

  const onOpenChange = useCallback(() => {
    navigate(`/organisations/${organisationId}/users`);
  }, [organisationId]);

  const onSubmit = async (data: InviteSchemaType) => {
    try {
      if (!organisationId) return;
      await client?.frontierServiceCreateOrganizationInvitation(
        organisationId,
        {
          user_ids: data?.emails,
          group_ids: data?.team,
          role_ids: data?.type,
        }
      );
      toast.success("Members added");
      navigate(`/organisations/${organisationId}/users`);
    } catch (err: unknown) {
      if (err instanceof Response && err?.status === 400) {
        toast.error("Bad Request", {
          description: (err as HttpResponse)?.error?.message,
        });
      } else {
        toast.error("Something went wrong", {
          description: (err as Error).message,
        });
      }
    }
  };

  useEffect(() => {
    async function getInformation(organisationId?: string) {
      try {
        setIsDataLoading(true);

        if (!organisationId) return;
        const [orgRolesResp, allRolesResp, groupsResp] = await Promise.all([
          client?.frontierServiceListOrganizationRoles(organisationId, {
            scopes: [PERMISSIONS.OrganizationNamespace],
          }),
          client?.frontierServiceListRoles({
            scopes: [PERMISSIONS.OrganizationNamespace],
          }),
          client?.frontierServiceListOrganizationGroups(organisationId),
        ]);
        setRoles([
          ...(orgRolesResp?.data?.roles || []),
          ...(allRolesResp?.data?.roles || []),
        ]);
        setGroups(groupsResp?.data?.groups || []);
      } catch (err) {
        console.error(err);
      } finally {
        setIsDataLoading(false);
      }
    }
    getInformation(organisationId);
  }, [client, organisationId]);

  const isSubmitting = methods?.formState?.isSubmitting;
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
              {isDataLoading ? (
                <Skeleton />
              ) : (
                <CustomFieldName
                  name="emails"
                  register={methods.register}
                  control={methods.control}
                  variant="textarea"
                  style={styles.textarea}
                  placeholder="Enter comma separated emails like abc@domain.com, bcd@domain.com"
                />
              )}
              {isDataLoading ? (
                <Skeleton />
              ) : (
                <CustomFieldName
                  name="type"
                  variant="select"
                  register={methods.register}
                  control={methods.control}
                  options={roles.map((r) => ({
                    label: r.title || "",
                    value: r?.id,
                  }))}
                />
              )}
              {isDataLoading ? (
                <Skeleton />
              ) : (
                <CustomFieldName
                  name="team"
                  variant="select"
                  register={methods.register}
                  control={methods.control}
                  options={groups.map((g) => ({
                    label: g.title || "",
                    value: g?.id,
                  }))}
                />
              )}
            </Flex>
            <SheetFooter>
              <Button
                type="submit"
                variant="primary"
                size={"medium"}
                disabled={isSubmitting}
              >
                <Text
                  style={{
                    color: "var(--foreground-inverted)",
                  }}
                >
                  {isSubmitting ? "Inviting..." : "Invite users"}
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
    width: "100%",
    backgroundColor: "var(--background-base)",
    border: "0.5px solid var(--border-base)",
    boxShadow: "var(--shadow-xs)",
    borderRadius: "var(--br-4)",
    color: "var(--foreground-base)",
    resize: "vertical",
  },
};
