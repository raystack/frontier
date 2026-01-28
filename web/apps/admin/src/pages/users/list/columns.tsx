import {
  Avatar,
  DataTableColumnDef,
  EmptyFilterValue,
  Flex,
  getAvatarColor,
  Text,
} from "@raystack/apsara";
import dayjs from "dayjs";
import styles from "./list.module.css";
import { getUserName, USER_STATES, UserState } from "../util";
import { User } from "@raystack/proton/frontier";
import {
  isNullTimestamp,
  TimeStamp,
  timestampToDate,
} from "~/utils/connect-timestamp";

interface getColumnsOptions {
  groupCountMap: Record<string, Record<string, number>>;
}

export const getColumns = ({
  groupCountMap,
}: getColumnsOptions): DataTableColumnDef<User, unknown>[] => {
  return [
    {
      accessorKey: "title",
      header: "Name",
      classNames: {
        cell: styles["name-column"],
        header: styles["name-column"],
      },
      cell: ({ row }) => {
        const avatarColor = getAvatarColor(row?.original?.id || "");
        const name = getUserName(row.original);
        return (
          <Flex gap={4} align="center">
            <Avatar
              src={row.original.avatar}
              fallback={name?.[0]?.toUpperCase()}
              color={avatarColor}
            />
            <Text>{name}</Text>
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
      accessorKey: "createdAt",
      header: "Joined on",
      filterType: "date",
      cell: ({ getValue }) => {
        const value = getValue() as TimeStamp;
        const date = isNullTimestamp(value)
          ? "-"
          : dayjs(timestampToDate(value)).format("YYYY-MM-DD");
        return <Text>{date}</Text>;
      },
      enableHiding: true,
      enableSorting: true,
    },
    {
      accessorKey: "state",
      header: "Status",
      cell: ({ getValue }) => {
        return USER_STATES?.[getValue() as UserState] ?? "-";
      },
      filterType: "select",
      filterOptions: Object.entries(USER_STATES).map(([value, label]) => ({
        value: value === "" ? EmptyFilterValue : value,
        label,
      })),
      enableColumnFilter: true,
      enableHiding: true,
      enableSorting: true,
      enableGrouping: true,
      showGroupCount: true,
      groupCountMap: groupCountMap["state"] || {},
      groupLabelsMap: USER_STATES,
    },
  ];
};
