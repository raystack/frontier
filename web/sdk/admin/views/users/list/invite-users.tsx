import { useMemo, useState, useEffect } from "react";
import {
  Button,
  Dialog,
  Flex,
  Label,
  Select,
  Text,
  TextArea,
  toastManager,
} from "@raystack/apsara";
import { PlusIcon } from "@radix-ui/react-icons";
import * as z from "zod";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { SCOPES, DEFAULT_ROLES } from "../../../utils/constants";
import styles from "./invite-users.module.css";
import Skeleton from "react-loading-skeleton";
import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  AdminServiceQueries,
  CreateOrganizationInvitationResponse,
  SearchOrganizationsRequestSchema,
  FrontierServiceQueries,
  ListRolesRequestSchema,
} from "@raystack/proton/frontier";
import {create} from "@bufbuild/protobuf";
import { handleConnectError } from "~/utils/error";
import { useTerminology } from "../../../hooks/useTerminology";

const inviteSchema = z.object({
  role: z.string(),
  organizationId: z.string(),
  emails: z
    .string()
    .transform(value => value.split(",").map(str => str.trim()))
    .pipe(z.array(z.string().email())),
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
    create(SearchOrganizationsRequestSchema, {query: {}}),
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
              <Flex direction="column" gap={2}>
                <Label className={styles["invite-users-dialog-label"]}>
                  Emails
                </Label>
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
                {(errors?.emails?.message || errors?.emails?.length) && (
                  <Text size="mini" className={styles["form-error-message"]}>
                    {errors?.emails?.message || errors?.emails?.[0]?.message}
                  </Text>
                )}
              </Flex>

              <Flex direction="column" gap={2}>
                <Label className={styles["invite-users-dialog-label"]}>
                  Invite as
                </Label>
                <Controller
                  name="role"
                  defaultValue={defaultRoleId}
                  disabled={isRolesLoading}
                  control={control}
                  render={({ field, fieldState: { error } }) => {
                    const { ref, ...rest } = field;
                    if (isRolesLoading) return <Skeleton height={33} />;
                    return (
                      <>
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
                        {error && (
                          <Text
                            size="mini"
                            className={styles["form-error-message"]}>
                            {error?.message}
                          </Text>
                        )}
                      </>
                    );
                  }}
                />
              </Flex>

              <Flex direction="column" gap={2}>
                <Label className={styles["invite-users-dialog-label"]}>
                  {t.organization({ case: "capital" })}
                </Label>
                <Controller
                  name="organizationId"
                  disabled={isOrganizationsLoading}
                  control={control}
                  render={({ field, fieldState: { error } }) => {
                    const { ref, ...rest } = field;
                    if (isOrganizationsLoading) return <Skeleton height={33} />;
                    return (
                      <>
                        <Select
                          {...rest}
                          onValueChange={value => field.onChange(value)}>
                          <Select.Trigger ref={ref}>
                            <Select.Value
                              placeholder={`Select ${t.organization({ case: "capital" })}`}
                            />
                          </Select.Trigger>
                          <Select.Content
                            style={{
                              maxHeight: 280,
                              overflowY: "auto",
                            }}>
                            {organizations?.map(org => (
                              <Select.Item key={org.id} value={org.id ?? ""}>
                                {org.name}
                              </Select.Item>
                            ))}
                          </Select.Content>
                        </Select>
                        {error && (
                          <Text
                            size="mini"
                            className={styles["form-error-message"]}>
                            {error?.message}
                          </Text>
                        )}
                      </>
                    );
                  }}
                />
              </Flex>
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Button
              data-test-id="users-list-invite-user-submit-btn"
              type="submit"
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
