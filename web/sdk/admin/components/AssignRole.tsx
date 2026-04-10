import {
  Button,
  Checkbox,
  Dialog,
  Flex,
  Label,
  Text,
  toast,
} from "@raystack/apsara";
import type {
  SearchOrganizationUsersResponse_OrganizationUser,
  Role,
} from "@raystack/proton/frontier";
import {
  FrontierServiceQueries,
  SetOrganizationMemberRoleRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
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
    formState: { isSubmitting, errors },
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

  function onCheckedChange(value: boolean | string, roleId?: string) {
    if (!roleId) return;
    if (value) {
      setValue("roleId", roleId);
    }
  }

  const onSubmit = async (data: FormData) => {
    try {
      if (data.roleId === currentRoleId) {
        onClose();
        return;
      }

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

      toast.success("Role assigned successfully");
    } catch (error) {
      toast.error("Failed to assign role");
      console.error(error);
    }
  };

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content width={400}>
        <Dialog.Header>
          <Dialog.Title>Assign Role</Dialog.Title>
          <Dialog.CloseButton data-test-id="assign-role-close-button" />
        </Dialog.Header>
        <form onSubmit={handleSubmit(onSubmit)} noValidate>
          <Dialog.Body>
            <Flex direction="column" gap={7}>
              <Text variant="secondary">
                Taking this action may result in changes in the role which might
                lead to changes in access of the user.
              </Text>
              <div role="group" aria-labelledby="roles-group">
                <Flex direction="column" gap={4}>
                  {roles.map((role) => {
                    const htmlId = `role-${role.id}`;
                    const checked = selectedRoleId === role.id;
                    return (
                      <Flex gap={3} key={role.id}>
                        <Checkbox
                          id={htmlId}
                          data-test-id={`role-checkbox-${role.id}`}
                          checked={checked}
                          onCheckedChange={(value) =>
                            onCheckedChange(value, role.id)
                          }
                        />
                        <Label htmlFor={htmlId}>{role.title}</Label>
                      </Flex>
                    );
                  })}
                  {errors.roleId && (
                    <Text variant="danger">{errors.roleId.message}</Text>
                  )}
                </Flex>
              </div>
            </Flex>
          </Dialog.Body>
          <Dialog.Footer>
            <Dialog.Close asChild>
              <Button
                type="button"
                variant="outline"
                color="neutral"
                data-test-id="assign-role-cancel-button"
              >
                Cancel
              </Button>
            </Dialog.Close>
            <Button
              type="submit"
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
