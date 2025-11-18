import { Text, DropdownMenu, Skeleton } from "@raystack/apsara";
import styles from "./side-panel.module.css";
import { useMemo, useState } from "react";
import {
  type SearchUserOrganizationsResponse_UserOrganization,
  SearchOrganizationUsersResponse_OrganizationUserSchema,
  type Role,
  FrontierServiceQueries,
  ListRolesRequestSchema,
  ListOrganizationRolesRequestSchema,
} from "@raystack/proton/frontier";
import { create } from "@bufbuild/protobuf";
import { useQuery } from "@connectrpc/connect-query";
import { SCOPES } from "~/utils/constants";
import { AssignRole } from "~/components/assign-role";
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

  const { data: defaultRoles = [], isLoading: isDefaultRolesLoading, error: defaultRolesError } = useQuery(
    FrontierServiceQueries.listRoles,
    create(ListRolesRequestSchema, { scopes: [SCOPES.ORG] }),
    {
      select: (data) => data?.roles || [],
    }
  );

  const { data: organizationRoles = [], isLoading: isOrgRolesLoading, error: orgRolesError } = useQuery(
    FrontierServiceQueries.listOrganizationRoles,
    create(ListOrganizationRolesRequestSchema, {
      orgId: data?.orgId || "",
      scopes: [SCOPES.ORG],
    }),
    {
      enabled: !!data?.orgId,
      select: (data) => data?.roles || [],
    }
  );

  // Log errors if they occur
  if (defaultRolesError) {
    console.error("Failed to fetch default roles:", defaultRolesError);
  }
  if (orgRolesError) {
    console.error("Failed to fetch organization roles:", orgRolesError);
  }

  const roles = useMemo(
    () => [...defaultRoles, ...organizationRoles],
    [defaultRoles, organizationRoles]
  );

  const isLoading = isDefaultRolesLoading || isOrgRolesLoading;

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
      <DropdownMenu>
        <DropdownMenu.Trigger className={styles["dropdown-menu-trigger"]}>
          <Text className={styles["text-overflow"]} as="p">
            {data?.roleTitles?.join(", ") ?? "-"}
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
