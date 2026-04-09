import {
  Button,
  Dialog,
  Flex,
  Label,
  Radio,
  Text,
  toast,
} from "@raystack/apsara";
import styles from "./members.module.css";
import type {
  SearchProjectUsersResponse_ProjectUser,
  Role,
} from "@raystack/proton/frontier";
import {
  FrontierServiceQueries,
  SetProjectMemberRoleRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useMutation } from "@connectrpc/connect-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

interface AssignRoleProps {
  projectId: string;
  roles: Role[];
  user?: SearchProjectUsersResponse_ProjectUser;
  onRoleUpdate: (user: SearchProjectUsersResponse_ProjectUser) => void;
  onClose: () => void;
}

const formSchema = z.object({
  roleId: z.string().min(1, "A role must be selected"),
});

type FormData = z.infer<typeof formSchema>;

export const AssignRole = ({
  roles = [],
  user,
  projectId,
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

  const { mutateAsync: setProjectMemberRole } = useMutation(
    FrontierServiceQueries.setProjectMemberRole,
  );

  const selectedRoleId = watch("roleId");

  const onSubmit = async (data: FormData) => {
    try {
      await setProjectMemberRole(
        create(SetProjectMemberRoleRequestSchema, {
          projectId,
          principalId: user?.id || "",
          principalType: "app/user",
          roleId: data.roleId,
        }),
      );

      if (onRoleUpdate) {
        onRoleUpdate({
          ...user,
          roleIds: [data.roleId],
        } as SearchProjectUsersResponse_ProjectUser);
      }

      toast.success("Role assigned successfully");
    } catch (error) {
      toast.error("Failed to assign role");
      console.error(error);
    }
  };

  return (
    <Dialog open onOpenChange={onClose}>
      <Dialog.Content
        width={400}
        overlayClassName={styles["action-dialog-overlay"]}
        className={styles["action-dialog-content"]}
      >
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
              <div role="radiogroup" aria-labelledby="roles-group">
                <Flex direction="column" gap={4}>
                  {roles.map((role) => {
                    const htmlId = `role-${role.id}`;
                    return (
                      <Flex gap={3} key={role.id}>
                        <Radio
                          id={htmlId}
                          data-test-id={`role-radio-${role.id}`}
                          checked={selectedRoleId === role.id}
                          onCheckedChange={() => setValue("roleId", role.id || "")}
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
