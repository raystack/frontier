import { SearchProjectUsersResponseProjectUser } from "~/api/frontier";
import styles from "./members.module.css";
import {
  Avatar,
  DataTableColumnDef,
  DropdownMenu,
  Flex,
  getAvatarColor,
  Text,
} from "@raystack/apsara/v1";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import { V1Beta1Role } from "@raystack/frontier";

interface getColumnsOptions {
  roles: V1Beta1Role[];
  handleAssignRoleAction: (user: SearchProjectUsersResponseProjectUser) => void;
  handleRemoveAction: (user: SearchProjectUsersResponseProjectUser) => void;
}

export const getColumns = ({
  handleAssignRoleAction,
  handleRemoveAction,
  roles = [],
}: getColumnsOptions): DataTableColumnDef<
  SearchProjectUsersResponseProjectUser,
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
      accessorKey: "role_ids",
      header: "Role",
      cell: ({ getValue }) => {
        const ids = getValue() as string[];
        return <Text>{ids.map((id) => roleMap[id]).join(", ")}</Text>;
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
        return (
          <DropdownMenu>
            <DropdownMenu.Trigger asChild>
              <DotsHorizontalIcon />
            </DropdownMenu.Trigger>
            <DropdownMenu.Content
              className={styles["table-action-dropdown"]}
              align="end"
            >
              <DropdownMenu.Item
                onSelect={() => handleAssignRoleAction(row.original)}
              >
                Assign role...
              </DropdownMenu.Item>
              <DropdownMenu.Item
                onSelect={() => handleRemoveAction(row.original)}
              >
                Remove user...
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu>
        );
      },
    },
  ];
};
