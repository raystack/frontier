import {
  DataTableColumnDef,
  getAvatarColor,
  Avatar,
  Flex,
  Text,
  Menu,
  IconButton,
  AlertDialog,
} from "@raystack/apsara";
import type {
  SearchOrganizationUsersResponse_OrganizationUser,
  Role,
} from "@raystack/proton/frontier";
import type { UpdateRolePayload } from "./update-role";
import styles from "./members.module.css";
import { DotsHorizontalIcon, UpdateIcon } from "@radix-ui/react-icons";
import { DeleteIcon } from "~/admin/assets/icons/DeleteIcon";
import { formatTimestamp, TimeStamp } from "~/admin/utils/connect-timestamp";
import { formatRoleTitle } from "~/admin/utils/helper";

const MemberStates = {
  enabled: "Active",
  disabled: "Suspended",
};

interface getColumnsOptions {
  roles: Role[];
  memberCount: number;
  updateRoleHandle: ReturnType<
    typeof AlertDialog.createHandle<UpdateRolePayload>
  >;
  handleRemoveMemberAction: (
    user: SearchOrganizationUsersResponse_OrganizationUser,
  ) => void;
}

export const getColumns = ({
  roles = [],
  memberCount,
  updateRoleHandle,
  handleRemoveMemberAction,
}: getColumnsOptions): DataTableColumnDef<
  SearchOrganizationUsersResponse_OrganizationUser,
  unknown
>[] => {
  const roleMap = roles.reduce(
    (acc, role) => {
      const id = role?.id ?? "";
      acc[id] = formatRoleTitle(role.title);
      return acc;
    },
    {} as Record<string, string>,
  );

  return [
    {
      accessorKey: "title",
      header: "Name",
      classNames: {
        cell: styles["name-column"],
        header: styles["name-column"],
      },
      cell: ({ row }) => {
        const nameInitial =
          row.original.title?.[0] || row?.original?.email?.[0];
        const avatarColor = getAvatarColor(row?.original?.id || "");
        return (
          <Flex gap={4} align="center">
            <Avatar
              src={row.original.avatar}
              fallback={nameInitial}
              color={avatarColor}
            />
            <Text>{row.original.title || "-"}</Text>
          </Flex>
        );
      },
      enableColumnFilter: true,
      enableSorting: true,
    },
    {
      accessorKey: "email",
      header: "Email",
      cell: ({ getValue }) => {
        return getValue();
      },
      enableColumnFilter: true,
    },
    {
      accessorKey: "roleIds",
      header: "Role",
      cell: ({ getValue }) => {
        const roleIds = getValue() as string[];
        return roleIds.map((id) => roleMap[id] || "-").join(", ");
      },
      enableColumnFilter: true,
      filterType: "select",
      filterOptions: roles.map((role) => ({
        label: role.title,
        value: role.id,
      })),
    },
    {
      accessorKey: "state",
      header: "Status",
      cell: ({ getValue }) => {
        return MemberStates[getValue() as keyof typeof MemberStates];
      },
      enableColumnFilter: true,
      filterType: "select",
      filterOptions: Object.entries(MemberStates).map(([key, value]) => ({
        label: value,
        value: key,
      })),
      enableHiding: true,
    },
    {
      accessorKey: "orgJoinedAt",
      header: "Joined On",
      cell: ({ getValue }) => (
        <Text>{formatTimestamp(getValue() as TimeStamp)}</Text>
      ),
      enableSorting: true,
      enableHiding: true,
      // enableColumnFilter: true,
      // filterType: "date",
    },
    {
      accessorKey: "id",
      header: "",
      classNames: {
        header: styles["table-action-column"],
        cell: styles["table-action-column"],
      },
      cell: ({ row }) => {
        // The last remaining member of an organization cannot be removed.
        const canRemoveMember = memberCount > 1;
        const userRoleIds = row.original.roleIds || [];
        // Only offer roles the member doesn't already have.
        const excludedRoles = roles.filter(
          (role) => role.id && !userRoleIds.includes(role.id),
        );
        return (
          <Menu>
            <Menu.Trigger
              render={
                <IconButton size={3} data-test-id="admin-members-action-menu">
                  <DotsHorizontalIcon />
                </IconButton>
              }
            />
            <Menu.Content className={styles["table-action-dropdown"]}>
              {excludedRoles.map((role) => (
                <Menu.Item
                  key={role.id}
                  leadingIcon={<UpdateIcon />}
                  onClick={() =>
                    updateRoleHandle.openWithPayload({
                      user: row.original,
                      role,
                    })
                  }
                  data-test-id={`admin-assign-role-${role.name}-action`}
                >
                  Make {role.title}
                </Menu.Item>
              ))}
              {canRemoveMember && (
                <Menu.Item
                  leadingIcon={
                    <DeleteIcon
                      style={{
                        color: "var(--rs-color-foreground-danger-primary)",
                      }}
                    />
                  }
                  onClick={() => handleRemoveMemberAction(row.original)}
                  data-test-id="admin-remove-member-action"
                  style={{ color: "var(--rs-color-foreground-danger-primary)" }}
                >
                  Remove
                </Menu.Item>
              )}
            </Menu.Content>
          </Menu>
        );
      },
    },
  ];
};
