import { Text, DropdownMenu } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import { useCallback, useEffect, useState } from "react";
import Skeleton from "react-loading-skeleton";
import { api } from "~/api";
import {
  SearchUserOrganizationsResponseUserOrganization,
  V1Beta1Role,
} from "~/api/frontier";
import { SCOPES } from "~/utils/constants";
import { AssignRole } from "~/components/assign-role";
import { useUser } from "../user-context";
import { SuspendUser } from "./suspend-user";

interface MembershipDropdownProps {
  data?: SearchUserOrganizationsResponseUserOrganization;
  onReset?: () => void;
}

export const MembershipDropdown = ({
  data,
  onReset,
}: MembershipDropdownProps) => {
  const [isLoading, setIsLoading] = useState(false);
  const [roles, setRoles] = useState<V1Beta1Role[]>([]);
  const [isAssignRoleDialogOpen, setIsAssignRoleDialogOpen] = useState(false);
  const [isSuspendDialogOpen, setIsSuspendDialogOpen] = useState(false);
  const user = useUser();

  const fetchRoles = useCallback(async (orgId: string) => {
    try {
      setIsLoading(true);
      const [defaultRolesResponse, organizationRolesResponse] =
        await Promise.all([
          api?.frontierServiceListRoles({
            scopes: [SCOPES.ORG],
          }),
          api?.frontierServiceListOrganizationRoles(orgId, {
            scopes: [SCOPES.ORG],
          }),
        ]);
      const defaultRoles = defaultRolesResponse.data?.roles || [];
      const organizationRoles = organizationRolesResponse.data?.roles || [];
      const roles = [...defaultRoles, ...organizationRoles];
      setRoles(roles);
    } catch (error) {
      console.error(error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (data?.org_id) fetchRoles(data.org_id);
  }, [data?.org_id, fetchRoles]);

  const toggleAssignRoleDialog = () => {
    setIsAssignRoleDialogOpen(value => !value);
  };

  const toggleSuspendDialog = () => {
    setIsSuspendDialogOpen(value => !value);
  };

  const onRoleUpdate = () => {
    toggleAssignRoleDialog();
    onReset?.();
  };

  const onSuspend = () => {
    toggleSuspendDialog();
    onReset?.();
  };

  if (isLoading) {
    return <Skeleton height={30} />;
  }

  return (
    <>
      {isAssignRoleDialogOpen && data?.org_id && (
        <AssignRole
          roles={roles}
          user={{
            ...user,
            role_ids: data?.role_ids,
            role_titles: data?.role_titles,
            role_names: data?.role_names,
          }}
          organizationId={data.org_id}
          onRoleUpdate={onRoleUpdate}
          onClose={toggleAssignRoleDialog}
        />
      )}
      {isSuspendDialogOpen && user?.id && (
        <SuspendUser
          userId={user.id}
          onClose={toggleSuspendDialog}
          onSubmit={onSuspend}
        />
      )}
      <DropdownMenu>
        <DropdownMenu.Trigger className={styles["dropdown-menu-trigger"]}>
          <Text className={styles["text-overflow"]}>
            {data?.role_titles?.join(", ") ?? "-"}
          </Text>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content>
          <DropdownMenu.Item
            onClick={toggleAssignRoleDialog}
            data-test-id="admin-ui-user-details-assign-role">
            Assign role...
          </DropdownMenu.Item>
          {/* TODO: Removed for now */}
          {/* <DropdownMenu.Item
            onClick={toggleSuspendDialog}
            data-test-id="admin-ui-user-details-suspend-user">
            Suspend...
          </DropdownMenu.Item> */}
        </DropdownMenu.Content>
      </DropdownMenu>
    </>
  );
};
