import styles from "./tokens.module.css";
import { NULL_DATE } from "~/utils/constants";
import dayjs from "dayjs";
import {
  Avatar,
  DataTableColumnDef,
  Flex,
  getAvatarColor,
  Text,
} from "@raystack/apsara";
import { SearchOrganizationTokensResponseOrganizationToken } from "~/api/frontier";

export const getColumns = (): DataTableColumnDef<
  SearchOrganizationTokensResponseOrganizationToken,
  unknown
>[] => {
  return [
    {
      accessorKey: "created_at",
      header: "Date",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      cell: ({ getValue }) => {
        const value = getValue() as string;
        return value !== NULL_DATE ? dayjs(value).format("YYYY-MM-DD") : "-";
      },
      enableSorting: true,
      enableColumnFilter: true,
      filterType: "date",
    },
    {
      accessorKey: "amount",
      header: "Tokens",
      cell: ({ row, getValue }) => {
        const prefix = row.original.type === "credit" ? "+" : "-";
        const value = getValue() as number;
        return `${prefix}${value}`;
      },
      enableSorting: true,
      enableColumnFilter: true,
      filterType: "number",
    },
    {
      accessorKey: "description",
      header: "Events",
      cell: ({ getValue }) => {
        return getValue();
      },
      enableHiding: true,
    },
    {
      accessorKey: "user_id",
      header: "Member",
      cell: ({ row, getValue }) => {
        const user_id = (getValue() as string) || "";
        const title = row.original.user_title || user_id;
        const avatarColor = getAvatarColor(user_id);
        return (
          <Flex gap={4} align="center">
            <Avatar
              src={row.original.user_avatar}
              fallback={title?.[0]}
              color={avatarColor}
            />
            <Text>{title}</Text>
          </Flex>
        );
      },
      enableHiding: true,
    },
  ];
};
