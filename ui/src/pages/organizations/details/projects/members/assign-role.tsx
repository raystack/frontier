import {
  Button,
  Checkbox,
  Dialog,
  Flex,
  Label,
  Text,
  toast,
} from "@raystack/apsara";
import styles from "./members.module.css";
import { useCallback } from "react";
import type {
  SearchProjectUsersResponse_ProjectUser,
  Role,
} from "@raystack/proton/frontier";
import {
  FrontierServiceQueries,
  FrontierService,
  ListPoliciesRequestSchema,
  DeletePolicyRequestSchema,
  CreatePolicyRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useMutation, useTransport } from "@connectrpc/connect-query";
import { createClient } from "@connectrpc/connect";
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
  onClose,
}: AssignRoleProps) => {
  const transport = useTransport();

  const {
    handleSubmit,
    watch,
    setValue,
    formState: { isSubmitting, errors },
  } = useForm<FormData>({
    defaultValues: {
      roleIds: new Set(user?.roleIds || []),
    },
    resolver: zodResolver(formSchema),
  });

  const { mutateAsync: deletePolicy } = useMutation(
    FrontierServiceQueries.deletePolicy,
  );

  const { mutateAsync: createPolicy } = useMutation(
    FrontierServiceQueries.createPolicy,
  );

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
      const client = createClient(FrontierService, transport);
      const policiesResp = await client.listPolicies(
        create(ListPoliciesRequestSchema, {
          projectId: projectId,
          userId: user?.id,
        }),
      );
      const policies = policiesResp.policies || [];

      const removedRolesPolicies = policies.filter(
        (policy) => !(policy.roleId && data.roleIds.has(policy.roleId)),
      );
      await Promise.all(
        removedRolesPolicies.map((policy) =>
          deletePolicy(
            create(DeletePolicyRequestSchema, { id: policy.id || "" }),
          ),
        ),
      );

      const resource = `app/project:${projectId}`;
      const principal = `app/user:${user?.id}`;

      const assignedRolesArr = Array.from(data.roleIds);
      await Promise.all(
        assignedRolesArr.map((roleId) =>
          createPolicy(
            create(CreatePolicyRequestSchema, {
              body: {
                roleId,
                resource,
                principal,
              },
            }),
          ),
        ),
      );

      if (onRoleUpdate) {
        onRoleUpdate({
          ...user,
          roleIds: assignedRolesArr,
        } as SearchProjectUsersResponse_ProjectUser);
      }

      toast.success("Role assigned successfully");
    } catch (error) {
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
