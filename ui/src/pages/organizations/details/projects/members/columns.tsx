import { SearchProjectUsersResponseProjectUser } from "~/api/frontier";
import styles from "./members.module.css";
import {
  Avatar,
  DataTableColumnDef,
  DropdownMenu,
  Flex,
  Text,
} from "@raystack/apsara/v1";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";

interface getColumnsOptions {
  handleAssignRoleAction: (id: string) => void;
  handleRemoveAction: (id: string) => void;
}

export const getColumns = ({
  handleAssignRoleAction,
  handleRemoveAction,
}: getColumnsOptions): DataTableColumnDef<
  SearchProjectUsersResponseProjectUser,
  unknown
>[] => {
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
        return (
          <Flex gap={4} align="center">
            <Avatar src={row.original.avatar} fallback={nameInitial} />
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
      accessorKey: "role_titles",
      header: "Role",
      cell: ({ getValue }) => {
        return getValue();
      },
    },
    {
      accessorKey: "id",
      header: "",
      classNames: {
        header: styles["table-action-column"],
        cell: styles["table-action-column"],
      },
      cell: ({ getValue }) => {
        const id = getValue() as string;
        return (
          <DropdownMenu open>
            <DropdownMenu.Trigger asChild>
              <DotsHorizontalIcon />
            </DropdownMenu.Trigger>
            <DropdownMenu.Content
              className={styles["table-action-dropdown"]}
              align="end"
            >
              <DropdownMenu.Item onSelect={() => handleAssignRoleAction(id)}>
                Assign role...
              </DropdownMenu.Item>
              <DropdownMenu.Item onSelect={() => handleRemoveAction(id)}>
                Remove user...
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu>
        );
      },
    },
  ];
};
