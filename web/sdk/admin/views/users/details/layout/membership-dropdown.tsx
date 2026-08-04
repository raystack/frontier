import { Text, Menu, Skeleton } from "@raystack/apsara";
import styles from "./side-panel.module.css";
import { useMemo, useState } from "react";
import {
  type SearchUserOrganizationsResponse_UserOrganization,
  SearchOrganizationUsersResponse_OrganizationUserSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useOrganizationRoles } from "~/admin/hooks/useOrganizationRoles";
import { AssignRole } from "../../../../components/AssignRole";
import { useUser } from "../user-context";
import { SuspendUser } from "./suspend-user";

interface MembershipDropdownProps {
  data?: SearchUserOrganizationsResponse_UserOrganization;
  onReset?: () => void;
}

export const MembershipDropdown = ({
  data,
  onReset,
}: MembershipDropdownProps) => {
  const [isAssignRoleDialogOpen, setIsAssignRoleDialogOpen] = useState(false);
  const [isSuspendDialogOpen, setIsSuspendDialogOpen] = useState(false);
  const { user } = useUser();

  const { roles, isLoading } = useOrganizationRoles(data?.orgId);

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

  const memoizedUser = useMemo(
    () =>
      create(
        SearchOrganizationUsersResponse_OrganizationUserSchema,
        Object.assign(user ?? {}, {
          roleNames: data?.roleNames || [],
          roleTitles: data?.roleTitles || [],
          roleIds: data?.roleIds || [],
        }),
      ),
    [user, data?.roleNames, data?.roleTitles, data?.roleIds],
  );

  if (isLoading) {
    return <Skeleton height={32} />;
  }

  return (
    <>
      {isAssignRoleDialogOpen && data?.orgId && (
        <AssignRole
          roles={roles}
          user={memoizedUser}
          organizationId={data.orgId}
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
      <Menu>
        <Menu.Trigger className={styles["dropdown-menu-trigger"]}>
          <Text className={styles["text-overflow"]} render={<p />}>
            {data?.roleTitles?.join(", ") ?? "-"}
          </Text>
        </Menu.Trigger>
        <Menu.Content>
          <Menu.Item
            onClick={toggleAssignRoleDialog}
            data-test-id="admin-user-details-assign-role">
            Assign role...
          </Menu.Item>
          {/* TODO: Removed for now */}
          {/* <Menu.Item
            onClick={toggleSuspendDialog}
            data-test-id="admin-user-details-suspend-user">
            Suspend...
          </Menu.Item> */}
        </Menu.Content>
      </Menu>
    </>
  );
};
