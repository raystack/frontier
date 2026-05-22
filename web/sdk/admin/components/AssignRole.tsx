import {
  Button,
  Dialog,
  Flex,
  Label,
  Radio,
  Text,
  toastManager,
} from "@raystack/apsara-v1";
import type {
  SearchOrganizationUsersResponse_OrganizationUser,
  Role,
} from "@raystack/proton/frontier";
import {
  FrontierServiceQueries,
  SetOrganizationMemberRoleRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { ConnectError } from "@connectrpc/connect";
import { useMutation } from "@connectrpc/connect-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

interface AssignRoleProps {
  organizationId: string;
  roles: Role[];
  user?: SearchOrganizationUsersResponse_OrganizationUser;
  onRoleUpdate: () => void;
  onClose: () => void;
}

const formSchema = z.object({
  roleId: z.string().min(1, "A role must be selected"),
});

type FormData = z.infer<typeof formSchema>;

export const AssignRole = ({
  roles = [],
  user,
  organizationId,
  onRoleUpdate,
  onClose,
}: AssignRoleProps) => {
  const currentRoleId = user?.roleIds?.[0] || "";

  const {
    handleSubmit,
    watch,
    setValue,
    formState: { isSubmitting, errors, isDirty },
  } = useForm<FormData>({
    defaultValues: {
      roleId: currentRoleId,
    },
    resolver: zodResolver(formSchema),
  });

  const { mutateAsync: setMemberRole } = useMutation(
    FrontierServiceQueries.setOrganizationMemberRole,
  );

  const selectedRoleId = watch("roleId");

  const onSubmit = async (data: FormData) => {
    try {
      await setMemberRole(
        create(SetOrganizationMemberRoleRequestSchema, {
          orgId: organizationId,
          userId: user?.id,
          roleId: data.roleId,
        }),
      );

      if (onRoleUpdate) {
        onRoleUpdate();
      }

      toastManager.add({
        title: "Role assigned successfully",
        type: "success",
      });
    } catch (error) {
      toastManager.add({
        title: "Failed to assign role",
        description: error instanceof ConnectError ? error.rawMessage : undefined,
        type: "error",
      });
      console.error(error);
    }
  };

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>Assign Role</Dialog.Title>
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)} noValidate>
          <Dialog.Body>
            <Flex direction="column" gap={7}>
              <Text variant="secondary">
                Taking this action may result in changes in the role which might
                lead to changes in access of the user.
              </Text>
              <Radio.Group
                value={selectedRoleId}
                onValueChange={(value) =>
                  setValue("roleId", value as string, { shouldDirty: true })
                }
              >
                <Flex direction="column" gap={4}>
                  {roles.map((role) => {
                    const htmlId = `role-${role.id}`;
                    return (
                      <Flex gap={3} key={role.id}>
                        <Radio
                          id={htmlId}
                          value={role.id || ""}
                          data-test-id={`role-radio-${role.id}`}
                        />
                        <Label htmlFor={htmlId}>{role.title}</Label>
                      </Flex>
                    );
                  })}
                  {errors.roleId && (
                    <Text variant="danger">{errors.roleId.message}</Text>
                  )}
                </Flex>
              </Radio.Group>
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Dialog.Close
              render={
                <Button
                  type="button"
                  variant="outline"
                  color="neutral"
                  data-test-id="assign-role-cancel-button"
                >
                  Cancel
                </Button>
              }
            />
            <Button
              type="submit"
              disabled={!isDirty}
              data-test-id="assign-role-update-button"
              loading={isSubmitting}
              loaderText="Updating..."
            >
              Update
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
