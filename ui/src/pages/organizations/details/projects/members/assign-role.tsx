import {
  Button,
  Checkbox,
  Dialog,
  Flex,
  Label,
  Text,
  toast,
} from "@raystack/apsara/v1";
import styles from "./members.module.css";
import { useCallback } from "react";
import {
  SearchProjectUsersResponseProjectUser,
  V1Beta1Role,
} from "~/api/frontier";
import { api } from "~/api";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

interface AssignRoleProps {
  projectId: string;
  roles: V1Beta1Role[];
  user?: SearchProjectUsersResponseProjectUser;
  onRoleUpdate: (user: SearchProjectUsersResponseProjectUser) => void;
}

const formSchema = z.object({
  roleIds: z.instanceof(Set<string>).refine((set) => set.size > 0, {
    message: "At least one role must be selected",
  }),
});

type FormData = z.infer<typeof formSchema>;

export const AssignRole = ({
  roles = [],
  user,
  projectId,
  onRoleUpdate,
}: AssignRoleProps) => {
  const {
    handleSubmit,
    watch,
    setValue,
    formState: { isSubmitting, errors },
  } = useForm<FormData>({
    defaultValues: {
      roleIds: new Set(user?.role_ids || []),
    },
    resolver: zodResolver(formSchema),
  });

  const roleIds = watch("roleIds");

  function onCheckedChange(value: boolean | string, roleId?: string) {
    if (!roleId) return;
    const currentRoles = new Set(roleIds);

    if (value) {
      currentRoles.add(roleId);
    } else {
      currentRoles.delete(roleId);
    }

    setValue("roleIds", currentRoles);
  }

  const checkRole = useCallback(
    (roleId?: string) => {
      if (!roleId) return false;
      return roleIds?.has(roleId) || false;
    },
    [roleIds],
  );

  const onSubmit = async (data: FormData) => {
    try {
      const policiesResp = await api?.frontierServiceListPolicies({
        project_id: projectId,
        user_id: user?.id,
      });
      const policies = policiesResp?.data?.policies || [];

      const removedRolesPolicies = policies.filter(
        (policy) => !(policy.role_id && data.roleIds.has(policy.role_id)),
      );
      await Promise.all(
        removedRolesPolicies.map((policy) =>
          api?.frontierServiceDeletePolicy(policy.id as string),
        ),
      );

      const resource = `app/project:${projectId}`;
      const principal = `app/user:${user?.id}`;

      const assignedRolesArr = Array.from(data.roleIds);
      await Promise.all(
        assignedRolesArr.map((role_id) =>
          api?.frontierServiceCreatePolicy({
            role_id,
            resource: resource,
            principal: principal,
          }),
        ),
      );

      if (onRoleUpdate) {
        onRoleUpdate({
          ...user,
          role_ids: assignedRolesArr,
        });
      }

      toast.success("Role assigned successfully");
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <Dialog open>
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
              <div role="group" aria-labelledby="roles-group">
                <Flex direction="column" gap={4}>
                  {roles.map((role) => {
                    const htmlId = `role-${role.id}`;
                    const checked = checkRole(role.id);
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
                  {errors.roleIds && (
                    <Text variant="danger">{errors.roleIds.message}</Text>
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
