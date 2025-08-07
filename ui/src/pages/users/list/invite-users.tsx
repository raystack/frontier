import { useMemo, useState, useEffect } from "react";
import {
  Button,
  Dialog,
  Flex,
  Label,
  Select,
  Text,
  TextArea,
  toast,
} from "@raystack/apsara";
import { PlusIcon } from "@radix-ui/react-icons";
import { AxiosError } from "axios";
import * as z from "zod";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { V1Beta1Role } from "@raystack/frontier";
import { api } from "~/api";
import { SCOPES, DEFAULT_ROLES } from "~/utils/constants";
import styles from "./invite-users.module.css";
import { useAppContext } from "~/contexts/App";
import Skeleton from "react-loading-skeleton";

const inviteSchema = z.object({
  role: z.string(),
  organizationId: z.string(),
  emails: z
    .string()
    .transform((value) => value.split(",").map((str) => str.trim()))
    .pipe(z.array(z.string().email())),
});

type InviteSchemaType = z.infer<typeof inviteSchema>;

const getDefaultRoles = async () => {
  const response = await api.frontierServiceListRoles({
    scopes: [SCOPES.ORG],
  });
  return response.data?.roles ?? [];
};

export const InviteUser = () => {
  const [open, onOpenChange] = useState(false);
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [isRolesLoading, setIsRolesLoading] = useState(false);
  const { organizations, isLoading } = useAppContext();

  const defaultRoleId = useMemo(
    () => roles?.find((role) => role.name === DEFAULT_ROLES.ORG_VIEWER)?.id,
    [roles],
  );

  useEffect(() => {
    const fetchRoles = async () => {
      try {
        setIsRolesLoading(true);
        const data = await getDefaultRoles();
        setRoles(data);
      } catch (error) {
        console.error("Failed to fetch roles:", error);
      } finally {
        setIsRolesLoading(false);
      }
    };

    fetchRoles();
  }, []);

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

  const onSubmit = async (data: InviteSchemaType) => {
    try {
      if (!data.organizationId) return;
      await api?.frontierServiceCreateOrganizationInvitation(
        data.organizationId,
        {
          user_ids: data?.emails,
          role_ids: data?.role ? [data?.role] : [],
        },
      );
      onOpenChange(false);
      reset({ role: defaultRoleId });

      toast.success(`User${data?.emails?.length > 1 ? "s" : ""} invited`);
    } catch (err: unknown) {
      if (err instanceof AxiosError && err?.status === 400) {
        toast.error("Bad Request", {
          description: err?.response?.data?.error?.message,
        });
      } else {
        toast.error("Something went wrong", {
          description: (err as Error).message,
        });
      }
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <Dialog.Trigger asChild>
        <Button
          variant="text"
          color="neutral"
          leadingIcon={<PlusIcon />}
          data-test-id="users-list-invite-user-btn"
        >
          Invite User
        </Button>
      </Dialog.Trigger>
      <Dialog.Content width={600}>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Dialog.Header>
            <Dialog.Title>Invite user</Dialog.Title>
            <Dialog.CloseButton data-test-id="users-list-invite-user-close-btn" />
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
                  render={({ field }) => (
                    <TextArea
                      {...field}
                      // @ts-expect-error placeholder props not defined in TS
                      placeholder="abc@example.com, xyz@example.com"
                    />
                  )}
                />
                {(errors?.emails?.message || errors?.emails?.length) && (
                  <Text size={1} className={styles["form-error-message"]}>
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
                          onValueChange={(value) => field.onChange(value)}
                        >
                          <Select.Trigger ref={ref}>
                            <Select.Value placeholder="Select a Role" />
                          </Select.Trigger>
                          <Select.Content>
                            {roles?.map((role) => (
                              <Select.Item key={role.id} value={role.id ?? ""}>
                                {role.title}
                              </Select.Item>
                            ))}
                          </Select.Content>
                        </Select>
                        {error && (
                          <Text
                            size={1}
                            className={styles["form-error-message"]}
                          >
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
                  Organization
                </Label>
                <Controller
                  name="organizationId"
                  disabled={isLoading}
                  control={control}
                  render={({ field, fieldState: { error } }) => {
                    const { ref, ...rest } = field;
                    if (isLoading) return <Skeleton height={33} />;
                    return (
                      <>
                        <Select
                          {...rest}
                          onValueChange={(value) => field.onChange(value)}
                        >
                          <Select.Trigger ref={ref}>
                            <Select.Value placeholder="Select an Organization" />
                          </Select.Trigger>
                          <Select.Content
                            style={{
                              maxHeight: 280,
                              overflowY: "auto",
                            }}
                          >
                            {organizations?.map((org) => (
                              <Select.Item key={org.id} value={org.id ?? ""}>
                                {org.name}
                              </Select.Item>
                            ))}
                          </Select.Content>
                        </Select>
                        {error && (
                          <Text
                            size={1}
                            className={styles["form-error-message"]}
                          >
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
              loaderText="Sending..."
            >
              Send invite
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
