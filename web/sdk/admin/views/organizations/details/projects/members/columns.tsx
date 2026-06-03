import type {
  SearchProjectUsersResponse_ProjectUser,
  Role,
} from "@raystack/proton/frontier";
import styles from "./members.module.css";
import {
  Avatar,
  Menu,
  Flex,
  getAvatarColor,
  Text,
  AlertDialog,
} from "@raystack/apsara-v1";
import type { DataTableColumnDef } from "@raystack/apsara-v1";
import { DotsHorizontalIcon, UpdateIcon } from "@radix-ui/react-icons";
import { DeleteIcon } from "~/admin/assets/icons/DeleteIcon";
import type { UpdateRolePayload } from "./update-role";

interface getColumnsOptions {
  roles: Role[];
  updateRoleHandle: ReturnType<
    typeof AlertDialog.createHandle<UpdateRolePayload>
  >;
  handleRemoveAction: (user: SearchProjectUsersResponse_ProjectUser) => void;
}

export const getColumns = ({
  updateRoleHandle,
  handleRemoveAction,
  roles = [],
}: getColumnsOptions): DataTableColumnDef<
  SearchProjectUsersResponse_ProjectUser,
  unknown
>[] => {
  const roleMap = roles.reduce(
    (acc, role) => {
      const id = role.id || "";
      acc[id] = role.title || "";
      return acc;
    },
    {} as Record<string, string>,
  );

  return [
    {
      accessorKey: "title",
      header: "Title",
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
      enableSorting: false,
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
        const ids = (getValue() as string[]) || [];
        return <Text>{ids.map((id) => roleMap[id]).join(", ") || "-"}</Text>;
      },
    },
    {
      accessorKey: "id",
      header: "",
      classNames: {
        header: styles["table-action-column"],
        cell: styles["table-action-column"],
      },
      cell: ({ row }) => {
        const userRoleIds = row.original.roleIds || [];
        // Only offer roles the member doesn't already have.
        const excludedRoles = roles.filter(
          (role) => role.id && !userRoleIds.includes(role.id),
        );
        return (
          <Menu>
            <Menu.Trigger render={<DotsHorizontalIcon />} />
            <Menu.Content
              side="bottom"
              align="end"
              className={styles["table-action-dropdown"]}
              //  @ts-ignore
              portal={false}
            >
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
              <Menu.Item
                leadingIcon={
                  <DeleteIcon
                    style={{
                      color: "var(--rs-color-foreground-danger-primary)",
                    }}
                  />
                }
                onClick={() => handleRemoveAction(row.original)}
                data-test-id="admin-remove-user-action"
                style={{ color: "var(--rs-color-foreground-danger-primary)" }}
              >
                Remove user
              </Menu.Item>
            </Menu.Content>
          </Menu>
        );
      },
    },
  ];
};
