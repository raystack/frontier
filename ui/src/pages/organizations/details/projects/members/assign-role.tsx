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
import { useCallback, useState } from "react";
import {
  SearchProjectUsersResponseProjectUser,
  V1Beta1Role,
} from "~/api/frontier";
import { api } from "~/api";

interface AssignRoleProps {
  projectId: string;
  roles: V1Beta1Role[];
  user?: SearchProjectUsersResponseProjectUser;
  onRoleUpdate: (user: SearchProjectUsersResponseProjectUser) => void;
}

export const AssignRole = ({
  roles = [],
  user,
  projectId,
  onRoleUpdate,
}: AssignRoleProps) => {
  const [assignedRoles, setAssignedRoles] = useState<Set<string>>(
    new Set(user?.role_ids || []),
  );
  const [isSubmitting, setIsSubmitting] = useState(false);

  function onCheckedChange(value: boolean | string, roleId?: string) {
    setAssignedRoles((prev) => {
      if (!roleId) return prev;
      // new set is needed to rerender
      const next = new Set(prev);
      if (value) {
        next.add(roleId);
      } else {
        next.delete(roleId);
      }
      return next;
    });
  }

  const checkRole = useCallback(
    (roleId?: string) => {
      if (!roleId) return false;
      return assignedRoles?.has(roleId) || false;
    },
    [assignedRoles],
  );

  const onSubmit = async () => {
    try {
      setIsSubmitting(true);
      const policiesResp = await api?.frontierServiceListPolicies({
        project_id: projectId,
        user_id: user?.id,
      });
      const policies = policiesResp?.data?.policies || [];

      const removedRolesPolicies = policies.filter(
        (policy) => !(policy.role_id && assignedRoles.has(policy.role_id)),
      );
      await Promise.all(
        removedRolesPolicies.map((policy) =>
          api?.frontierServiceDeletePolicy(policy.id as string),
        ),
      );

      const resource = `app/project:${projectId}`;
      const principal = `app/user:${user?.id}`;

      const assignedRolesArr = Array.from(assignedRoles);
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
    } finally {
      setIsSubmitting(false);
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
        <Dialog.Body>
          <Flex direction="column" gap={7}>
            <Text variant="secondary">
              Taking this action may result in changes in the role which might
              lead to changes in access of the user.
            </Text>
            <Flex direction="column" gap={4}>
              {roles.map((role) => {
                const htmlId = `role-${role.id}`;
                const checked = checkRole(role.id);
                return (
                  <Flex gap={3} key={role.id}>
                    <Checkbox
                      id={htmlId}
                      checked={checked}
                      onCheckedChange={(value) =>
                        onCheckedChange(value, role.id)
                      }
                    />
                    <Label htmlFor={htmlId}>{role.title}</Label>
                  </Flex>
                );
              })}
            </Flex>
          </Flex>
        </Dialog.Body>
        <Dialog.Footer>
          <Dialog.Close asChild>
            <Button
              variant="outline"
              color="neutral"
              data-test-id="assign-role-cancel-button"
            >
              Cancel
            </Button>
          </Dialog.Close>
          <Button
            data-test-id="assign-role-update-button"
            onClick={onSubmit}
            loading={isSubmitting}
            loaderText="Updating..."
          >
            Update
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
