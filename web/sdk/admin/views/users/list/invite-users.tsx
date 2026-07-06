import { useMemo, useState, useEffect } from "react";
import {
  Button,
  Dialog,
  Field,
  Flex,
  Select,
  TextArea,
  Skeleton,
  toastManager,
} from "@raystack/apsara";
import { PlusIcon } from "@radix-ui/react-icons";
import * as z from "zod";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { SCOPES, DEFAULT_ROLES } from "../../../utils/constants";
import styles from "./invite-users.module.css";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  CreateOrganizationInvitationResponse,
  SearchOrganizationsRequestSchema,
  FrontierServiceQueries,
  ListRolesRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { handleConnectError } from "~/utils/error";
import { useTerminology } from "../../../hooks/useTerminology";

const inviteSchema = z.object({
  role: z.string().min(1, { message: "Role is required" }),
  organizationId: z.string().min(1, { message: "Organization is required" }),
  emails: z
    .string()
    .min(1, { message: "Email is required" })
    .transform(value =>
      value
        .split(",")
        .map(str => str.trim())
        .filter(str => str.length > 0)
    )
    .pipe(
      z
        .array(z.string().email({ message: "Enter valid email address(es)" }))
        .min(1, { message: "Email is required" })
    ),
});

type InviteSchemaType = z.infer<typeof inviteSchema>;

export const InviteUser = () => {
  const t = useTerminology();
  const [open, onOpenChange] = useState(false);

  const {
    data: organizations,
    isLoading: isOrganizationsLoading,
    error: organizationsError,
  } = useQuery(
    AdminServiceQueries.searchOrganizations,
    create(SearchOrganizationsRequestSchema, { query: {} }),
    {
      select: (data) => data?.organizations || [],
    }
  );

  const {
    data: roles,
    isLoading: isRolesLoading,
    error: rolesError,
  } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, { scopes: [SCOPES.ORG] }),
    {
      select: (data) => data?.roles || [],
    }
  );

  const isLoading = isRolesLoading || isOrganizationsLoading;

  useEffect(() => {
    if (organizationsError) {
      console.error("Failed to fetch organizations:", organizationsError);
    }
  }, [organizationsError]);

  useEffect(() => {
    if (rolesError) {
      console.error("Failed to fetch roles:", rolesError);
    }
  }, [rolesError]);

  const defaultRoleId = useMemo(
    () => roles?.find(role => role.name === DEFAULT_ROLES.ORG_VIEWER)?.id,
    [roles],
  );

  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    control,
    reset,
  } = useForm<InviteSchemaType>({
    resolver: zodResolver(inviteSchema),
    defaultValues: {
      role: defaultRoleId,
    },
  });

  const { mutateAsync: inviteUser } = useMutation(
    FrontierServiceQueries.createOrganizationInvitation,
    {
      onSuccess: (data: CreateOrganizationInvitationResponse) => {
        const invitedCount = data?.invitations?.length ?? 0;
        toastManager.add({
          title: `${t.user({ case: "capital", plural: invitedCount !== 1 })} invited`,
          type: "success",
        });
        reset({ role: defaultRoleId });
        onOpenChange(false);
      },
    },
  );

  const onSubmit = async (data: InviteSchemaType) => {
    if (!data.organizationId) return;
    try {
      await inviteUser({
        orgId: data.organizationId,
        userIds: data?.emails,
        roleIds: data?.role ? [data?.role] : [],
      });
    } catch (error) {
      handleConnectError(error, {
        AlreadyExists: () => toastManager.add({ title: 'Invitation already exists', type: "error" }),
        InvalidArgument: (err) => toastManager.add({ title: 'Invalid input', description: err.rawMessage, type: "error" }),
        PermissionDenied: () => toastManager.add({ title: "You don't have permission to perform this action", type: "error" }),
        Default: (err) => toastManager.add({ title: 'Something went wrong', description: err.rawMessage, type: "error" }),
      });
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Trigger
        render={
          <Button
            variant="text"
            color="neutral"
            leadingIcon={<PlusIcon />}
            data-test-id="users-list-invite-user-btn">
            Invite {t.user({ case: "capital" })}
          </Button>
        }
      />
      <Dialog.Content width={600}>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Header>
            <Dialog.Title>Invite {t.user({ case: "lower" })}</Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className={styles["invite-users-dialog-body"]}>
            <Flex direction="column" gap={7}>
              <Field
                label="Emails"
                error={errors?.emails?.message || errors?.emails?.[0]?.message}>
                {isLoading ? (
                  <Skeleton height="80px" />
                ) : (
                  <Controller
                    name="emails"
                    control={control}
                    render={({ field }) => {
                      const { value, ...rest } = field;
                      return (
                        <TextArea
                          {...rest}
                          value={Array.isArray(value) ? value.join(", ") : (value ?? "") as string}
                          placeholder="abc@example.com, xyz@example.com"
                          className={styles["invite-users-emails-textarea"]}
                        />
                      );
                    }}
                  />
                )}
              </Field>

              <Field label="Invite as" error={errors?.role?.message}>
                {isLoading ? (
                  <Skeleton height="36px" />
                ) : (
                  <Controller
                    name="role"
                    defaultValue={defaultRoleId}
                    control={control}
                    render={({ field }) => {
                      const { ref, ...rest } = field;
                      return (
                        <Select
                          {...rest}
                          onValueChange={value => field.onChange(value)}>
                          <Select.Trigger ref={ref}>
                            <Select.Value placeholder="Select a Role" />
                          </Select.Trigger>
                          <Select.Content>
                            {roles?.map(role => (
                              <Select.Item key={role.id} value={role.id ?? ""}>
                                {role.title}
                              </Select.Item>
                            ))}
                          </Select.Content>
                        </Select>
                      );
                    }}
                  />
                )}
              </Field>

              <Field
                label={t.organization({ case: "capital" })}
                error={errors?.organizationId?.message}>
                {isLoading ? (<Skeleton height="36px" />) : (<Controller
                  name="organizationId"
                  control={control}
                  render={({ field }) => {
                    const { ref, ...rest } = field;
                    return (
                      <Select
                        {...rest}
                        autocomplete
                        onValueChange={value => field.onChange(value)}>
                        <Select.Trigger ref={ref}>
                          <Select.Value
                            placeholder={`Select ${t.organization({ case: "capital" })}`}
                          />
                        </Select.Trigger>
                        <Select.Content
                          searchPlaceholder={`Search ${t.organization({ case: "lower" })}`}
                          style={{
                            maxHeight: 280,
                            overflowY: "auto",
                          }}>
                          {organizations?.map(org => (
                            <Select.Item key={org.id} value={org.id ?? ""}>
                              {org.title || org.name}
                            </Select.Item>
                          ))}
                        </Select.Content>
                      </Select>
                    );
                  }}
                />)}
              </Field>
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Button
              data-test-id="users-list-invite-user-submit-btn"
              type="submit"
              disabled={isLoading}
              loading={isSubmitting}
              loaderText="Sending...">
              Send invite
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
