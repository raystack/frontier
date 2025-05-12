import {
  Button,
  Dialog,
  Flex,
  Label,
  Select,
  Text,
  TextArea,
  toast,
} from "@raystack/apsara/v1";
import { useContext, useMemo } from "react";
import styles from "./layout.module.css";
import { OrganizationContext } from "../contexts/organization-context";
import * as z from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, FormProvider, useForm } from "react-hook-form";
import { AxiosError } from "axios";
import { api } from "~/api";
import { DEFAULT_INVITE_ROLE } from "../types";

const inviteSchema = z.object({
  role: z.string(),
  emails: z
    .string()
    .transform((value) => value.split(",").map((str) => str.trim()))
    .pipe(z.array(z.string().email())),
});

type InviteSchemaType = z.infer<typeof inviteSchema>;

interface InviteUsersDialogProps {
  onOpenChange: (open: boolean) => void;
}

export const InviteUsersDialog = ({ onOpenChange }: InviteUsersDialogProps) => {
  const { roles = [], organization } = useContext(OrganizationContext);
  const organizationId = organization?.id || "";

  const methods = useForm<InviteSchemaType>({
    resolver: zodResolver(inviteSchema),
    defaultValues: {},
  });

  const onSubmit = async (data: InviteSchemaType) => {
    try {
      if (!organizationId) return;
      await api?.frontierServiceCreateOrganizationInvitation(organizationId, {
        user_ids: data?.emails,
        role_ids: data?.role ? [data?.role] : [],
      });
      onOpenChange(false);
      toast.success("user invited");
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

  const defaultRoleId = useMemo(
    () => roles?.find((role) => role.name === DEFAULT_INVITE_ROLE)?.id,
    [roles],
  );

  const isSubmitting = methods?.formState?.isSubmitting;
  const errors = methods?.formState?.errors;
  return (
    <Dialog open onOpenChange={onOpenChange}>
      <Dialog.Content width={600}>
        <FormProvider {...methods}>
          <form onSubmit={methods.handleSubmit(onSubmit)}>
            <Dialog.Header>
              <Dialog.Title>Invite user</Dialog.Title>
              <Dialog.CloseButton data-test-id="invite-users-close-button" />
            </Dialog.Header>
            <Dialog.Body className={styles["invite-users-dialog-body"]}>
              <Flex direction="column" gap={7}>
                <Flex direction="column" gap={2}>
                  <Label className={styles["invite-users-dialog-label"]}>
                    Emails
                  </Label>
                  <Controller
                    name="emails"
                    control={methods.control}
                    render={({ field }) => (
                      <TextArea
                        {...field}
                        // @ts-expect-error placeholder props not defined in TS
                        placeholder="abc@example.com, xyz@example.com"
                      />
                    )}
                  />
                  {errors?.emails?.message || errors?.emails?.length ? (
                    <Text size={1} className={styles["form-error-message"]}>
                      {errors?.emails?.message || errors?.emails?.[0]?.message}
                    </Text>
                  ) : null}
                </Flex>

                <Flex direction="column" gap={2}>
                  <Label className={styles["invite-users-dialog-label"]}>
                    Role
                  </Label>
                  <Controller
                    name="role"
                    control={methods.control}
                    render={({ field }) => {
                      const { ref, ...rest } = field;
                      return (
                        <Select
                          {...rest}
                          onValueChange={(value: any) => field.onChange(value)}
                          defaultValue={defaultRoleId}
                        >
                          <Select.Trigger ref={ref}>
                            <Select.Value placeholder="Select Role" />
                          </Select.Trigger>
                          <Select.Content
                            className={
                              styles["invite-users-dialog-roles-content"]
                            }
                          >
                            {roles?.map((role) => (
                              <Select.Item key={role.id} value={role.id || ""}>
                                {role.title}
                              </Select.Item>
                            ))}
                          </Select.Content>
                        </Select>
                      );
                    }}
                  />

                  {errors?.role?.message && (
                    <Text size={1} className={styles["form-error-message"]}>
                      {errors?.role?.message}
                    </Text>
                  )}
                </Flex>
              </Flex>
            </Dialog.Body>
            <Dialog.Footer>
              <Button
                data-test-id="invite-users-invite-button"
                type="submit"
                loading={isSubmitting}
                loaderText="Inviting..."
              >
                Invite
              </Button>
            </Dialog.Footer>
          </form>
        </FormProvider>
      </Dialog.Content>
    </Dialog>
  );
};
