import { SearchProjectUsersResponseProjectUser } from "~/api/frontier";
import styles from "./members.module.css";
import { Avatar, DataTableColumnDef, Flex, Text } from "@raystack/apsara/v1";

export const getColumns = (): DataTableColumnDef<
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
  ];
};
