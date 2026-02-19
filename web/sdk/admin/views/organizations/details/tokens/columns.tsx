import styles from "./tokens.module.css";
import { NULL_DATE } from "../../../../utils/constants";
import dayjs from "dayjs";
import {
  Avatar,
  DataTableColumnDef,
  Flex,
  getAvatarColor,
  Text,
} from "@raystack/apsara";
import type {
  SearchOrganizationTokensResponse_OrganizationToken,
} from "@raystack/proton/frontier";
import {
  isNullTimestamp,
  TimeStamp,
  timestampToDate,
} from "../../../../utils/connect-timestamp";

export const getColumns = (): DataTableColumnDef<
  SearchOrganizationTokensResponse_OrganizationToken,
  unknown
>[] => {
  return [
    {
      accessorKey: "createdAt",
      header: "Date",
      classNames: {
        cell: styles["first-column"],
        header: styles["first-column"],
      },
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        const date = isNullTimestamp(value)
          ? "-"
          : dayjs(timestampToDate(value)).format("YYYY-MM-DD");
        return date;
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
      accessorKey: "userId",
      header: "Member",
      cell: ({ row, getValue }) => {
        const userId = (getValue() as string) || "";
        const title = row.original.userTitle || userId;
        const avatarColor = getAvatarColor(userId);
        return (
          <Flex gap={4} align="center">
            <Avatar
              src={row.original.userAvatar}
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
