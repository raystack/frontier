import { DataTableColumnDef } from "@raystack/apsara/v1";
import { SearchOrganizationUsersResponseOrganizationUser } from "~/api/frontier";
import styles from "./members.module.css";
import { Avatar, Flex, Text } from "@raystack/apsara/v1";
import dayjs from "dayjs";
import { NULL_DATE } from "~/utils/constants";

export const getColumns = (): DataTableColumnDef<
  SearchOrganizationUsersResponseOrganizationUser,
  unknown
>[] => {
  return [
    {
      accessorKey: "title",
      header: "Name",
      classNames: {
        cell: styles["name-column"],
        header: styles["name-column"],
      },
      cell: ({ row }) => {
        return (
          <Flex gap={4} align="center">
            <Avatar
              src={row.original.avatar}
              fallback={row.original.title?.[0]}
            />
            <Text>{row.original.title}</Text>
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
      enableColumnFilter: true,
    },
    {
      accessorKey: "state",
      header: "Status",
      cell: ({ getValue }) => {
        return getValue();
      },
      enableColumnFilter: true,
    },
    {
      accessorKey: "joined_at",
      header: "Joined On",
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
      },
      enableColumnFilter: true,
    },
  ];
};
